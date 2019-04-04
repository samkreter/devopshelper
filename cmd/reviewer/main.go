package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/samkreter/vstsautoreviewer/pkg/autoreviewer"
	"github.com/samkreter/vstsautoreviewer/pkg/config"
	"github.com/samkreter/vstsautoreviewer/pkg/types"

	vstsObj "github.com/samkreter/vsts-goclient/api/git"
	vsts "github.com/samkreter/vsts-goclient/client"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configFilePath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		log.Fatal("CONFIG_PATH not set")
	}

	conf, err := config.LoadConfig(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	aReviewers := make([]*autoreviewer.AutoReviewer, 0, len(conf.RepositoryInfos))
	for _, repoInfo := range conf.RepositoryInfos {
		aReviewer, err := getAutoReviewers(repoInfo, conf)
		if err != nil {
			log.Printf("ERROR: Failed to init reviewer for repo: %s/%s with err: %v", repoInfo.ProjectName, repoInfo.RepositoryName, err)
			continue
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

func getAutoReviewers(repoInfo *config.RepositoryInfo, conf *config.Config) (*autoreviewer.AutoReviewer, error) {
	vstsConfig := &vsts.Config{
		Token:          conf.Token,
		Username:       conf.Username,
		APIVersion:     conf.APIVersion,
		RepositoryName: repoInfo.RepositoryName,
		Project:        repoInfo.ProjectName,
		Instance:       conf.Instance,
	}

	vstsClient, err := vsts.NewClient(vstsConfig)
	if err != nil {
		return nil, err
	}

	filters := []autoreviewer.Filter{
		filterWIP,
		filterMasterBranchOnly,
	}

	reviewerTriggers := make([]autoreviewer.ReviwerTrigger, 0)

	slackTriggerPath, ok := os.LookupEnv("SLACK_TRIGGER_PATH")
	if ok {
		slackTrigger, err := autoreviewer.NewSlackTrigger(slackTriggerPath)
		if err != nil {
			log.Printf("ERROR: Failed to create slack trigger with error: %v", err)
		} else {
			reviewerTriggers = append(reviewerTriggers, slackTrigger)
			log.Println("Adding Slack Reviewer Trigger...")
		}
	}

	reviewerGroups, err := loadReviewerGroups(repoInfo.ReviewerFile, repoInfo.StatusFile)
	if err != nil {
		return nil, err
	}

	aReviewer, err := autoreviewer.NewAutoReviewer(vstsClient, conf.BotMaker, reviewerGroups, filters, reviewerTriggers)
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

func filterMasterBranchOnly(pr vstsObj.GitPullRequest) bool {
	if strings.EqualFold(pr.TargetRefName, "refs/heads/master") {
		return false
	}

	return true
}

func loadReviewerGroups(reviewerFile, statusFile string) (types.ReviewerGroups, error) {
	rawReviewerData, err := ioutil.ReadFile(reviewerFile)
	if err != nil {
		return nil, fmt.Errorf("Could not load %s", reviewerFile)
	}

	var reviewerGroups types.ReviewerGroups
	err = json.Unmarshal(rawReviewerData, &reviewerGroups)
	if err != nil {
		return nil, err
	}

	reviewerPoses := types.ReviewerPositions{}
	if _, err := os.Stat(statusFile); os.IsNotExist(err) {
		// Create the current pos file if it doesn't exist
		reviewerGroups.SavePositions(statusFile)
	}

	rawPosData, err := ioutil.ReadFile(statusFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read status file err: '%v'", err)
	}

	err = json.Unmarshal(rawPosData, &reviewerPoses)
	if err != nil {
		return nil, err
	}

	for index, reviewerGroup := range reviewerGroups {
		if pos, ok := reviewerPoses[reviewerGroup.Group]; ok {
			reviewerGroups[index].CurrentPos = pos
		}
	}

	return reviewerGroups, nil
}
