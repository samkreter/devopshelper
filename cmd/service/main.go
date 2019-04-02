package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/samkreter/vstsautoreviewer/pkg/autoreviewer"
	"github.com/samkreter/vstsautoreviewer/pkg/config"

	vstsObj "github.com/samkreter/vsts-goclient/api/git"
	vsts "github.com/samkreter/vsts-goclient/client"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var (
	configFilePath string
)

func main() {
	flag.StringVar(&configFilePath, "config-file", "", "filepath to the configuration file.")
	flag.Parse()

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

	aReviewer, err := autoreviewer.NewAutoReviewer(vstsClient, conf.BotMaker, repoInfo.ReviewerFile, repoInfo.StatusFile, filters, reviewerTriggers)
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
