package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"images/vstsAutoReviewer/autoreviewer"

	vstsObj "github.com/samkreter/vsts-goclient/api/git"
	vsts "github.com/samkreter/vsts-goclient/client"
)

var (
	defaultReviewerFile = "/configs/reviewers.json"
	defaultStatusFile   = "/configs/currentStatus.json"
)

// Config holds the configuration from the config file
type Config struct {
	Token           string           `json:"token"`
	Username        string           `json:"username"`
	APIVersion      string           `json:"apiVersion"`
	RepositoryInfos []repositoryInfo `json:"repositoryInfos"`
	Project         string           `json:"project"`
	Instance        string           `json:"instance"`
	BotMaker        string           `json:"botMaker"`
	ReviewerFile    string           `json:"reviewerFile"`
	StatusFile      string           `json:"statusFile"`
}

type repositoryInfo struct {
	RepositoryName string `json:"repositoryName"`
	ProjectName    string `json:"projectName"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configFilePath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		log.Fatal("CONFIG_PATH not set")
	}

	configFile, err := os.Open(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	defer configFile.Close()

	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatal(err)
	}

	if config.ReviewerFile == "" {
		config.ReviewerFile = defaultReviewerFile
	}

	if config.StatusFile == "" {
		config.StatusFile = defaultStatusFile
	}

	aReviewers := make([]*autoreviewer.AutoReviewer, 0, len(config.RepositoryInfos))
	for _, repoInfo := range config.RepositoryInfos {
		aReviewer, err := getAutoReviewers(repoInfo, config)
		if err != nil {
			log.Printf("ERROR: Failed to init reviewer for repo: %s/%s with err: %v", repoInfo.ProjectName, repoInfo.RepositoryName, err)
		}

		aReviewers = append(aReviewers, aReviewer)
	}

	for _, aReviewer := range aReviewers {
		log.Printf("Starting Reviewer for repo: %s\n", aReviewer.Repository)
		if err := aReviewer.Run(); err != nil {
			log.Printf("Failed to balance repo: %s with err: %v\n", aReviewer.Repository, err)
		}
		log.Printf("Finished Balancing Cycle for repo: %s\n", aReviewer.Repository)
	}

	log.Println("Finished Reviewing for all repositories")
}

func getAutoReviewers(repoInfo repositoryInfo, config Config) (*autoreviewer.AutoReviewer, error) {
	vstsConfig := &vsts.Config{
		Token:          config.Token,
		Username:       config.Username,
		APIVersion:     config.APIVersion,
		RepositoryName: repoInfo.RepositoryName,
		Project:        repoInfo.ProjectName,
		Instance:       config.Instance,
	}

	vstsClient, err := vsts.NewClient(vstsConfig)
	if err != nil {
		return nil, err
	}

	filters := []autoreviewer.Filter{
		filterWIP,
	}

	aReviewer, err := autoreviewer.NewAutoReviewer(vstsClient, config.BotMaker, config.ReviewerFile, config.StatusFile, filters)
	if err != nil {
		return nil, err
	}

	return aReviewer, nil
}

func filterWIP(pr vstsObj.GitPullRequest) bool {
	if strings.Contains(pr.Title, "WIP") {
		return true
	}

	return false
}
