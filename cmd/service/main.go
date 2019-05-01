package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/samkreter/vstsautoreviewer/pkg/autoreviewer"
	"github.com/samkreter/vstsautoreviewer/pkg/config"
	"github.com/samkreter/vstsautoreviewer/pkg/server"
	"github.com/samkreter/vstsautoreviewer/pkg/store"
	"github.com/samkreter/vstsautoreviewer/pkg/types"

	"github.com/samkreter/go-core/log"
	vstsObj "github.com/samkreter/vsts-goclient/api/git"
	vsts "github.com/samkreter/vsts-goclient/client"
)

const (
	defaultVSTSAPIVersion      = "5.0"
	defaultReviewerIntervalMin = 5
)

var (
	configFilePath string
	admins         string
	vstsToken      string
	vstsUsername   string
	serverAddr     string
	logLvl         string
	conf           = &config.Config{}
	mongoOptions   = &store.MongoStoreOptions{}
)

func main() {
	flag.StringVar(&configFilePath, "config-file", "", "filepath to the configuration file.")
	flag.StringVar(&serverAddr, "addr", "localhost:8080", "the address for the api server to listen on.")
	flag.StringVar(&logLvl, "log-level", "info", "the log level for the application")

	flag.StringVar(&conf.Token, "vsts-token", "", "vsts personal access token")
	flag.StringVar(&conf.Username, "vsts-username", "", "vsts username")

	flag.StringVar(&admins, "admins", "", "admins to be added to each repo, comma seperated")

	flag.StringVar(&mongoOptions.MongoURI, "mongo-uri", "", "connection string for the mongo database")
	flag.StringVar(&mongoOptions.RepositoryCollection, "mongo-repo-collection", "", "collection that stores the repositories")
	flag.StringVar(&mongoOptions.BaseGroupCollection, "mongo-basegroup-collection", "", "collection that stores the base groups")
	flag.StringVar(&mongoOptions.DBName, "mongo-dbname", "reviewerBot", "the mongo database to access")
	flag.BoolVar(&mongoOptions.UseSSL, "mongo-ssl", false, "use ssl when accessing mongo database")

	flag.StringVar(&conf.BotMaker, "botmaker-id", "b03f5f7f11d50a3a", "identifier for the bot's message")
	flag.StringVar(&conf.Instance, "vsts-instance", "msazure.visualstudio.com", "vsts instance")
	flag.StringVar(&conf.APIVersion, "vsts-apiversion", defaultVSTSAPIVersion, "vsts instance")

	reviewIntervalMin := flag.Int("review-interval", defaultReviewerIntervalMin, "number of minutes to wait to reviwer")
	flag.Parse()

	ctx := context.Background()
	logger := log.G(ctx)

	if err := log.SetLogLevel(logLvl); err != nil {
		logger.Errorf("failed to set log level to : '%s'", logLvl)
	}

	var err error
	if configFilePath != "" {
		conf, err = config.LoadConfig(configFilePath)
		if err != nil {
			logger.Fatal(err)
		}
	}

	adminsSplit := strings.Split(admins, ",")

	repoStore, err := store.NewMongoStore(mongoOptions)
	if err != nil {
		logger.Fatal(err)
	}

	vstsConfig := &vsts.Config{
		Token:      conf.Token,
		Username:   conf.Username,
		APIVersion: conf.APIVersion,
		Instance:   conf.Instance,
	}

	vstsClient, err := vsts.NewClient(vstsConfig)
	if err != nil {
		logger.Fatalf("failed to create vsts client with err: '%v'", err)
	}

	go func() {
		logger.Info("Starting Reviewer....")
		err = processReviewers(ctx, repoStore, conf)
		if err != nil {
			logger.Error(err)
		}
		for range time.NewTicker(time.Minute * time.Duration(*reviewIntervalMin)).C {
			err = processReviewers(ctx, repoStore, conf)
			if err != nil {
				logger.Error(err)
			}
		}
	}()

	s, err := server.NewServer(serverAddr, vstsClient, repoStore, adminsSplit)
	if err != nil {
		logger.Fatal(err)
	}

	// Run the apiserver
	s.Run()
}

func processReviewers(ctx context.Context, repoStore store.RepositoryStore, conf *config.Config) error {
	logger := log.G(ctx)
	repos, err := repoStore.GetAllRepositories(ctx)
	if err != nil {
		return err
	}

	aReviewers := make([]*autoreviewer.AutoReviewer, 0, len(repos))
	for _, repo := range repos {
		aReviewer, err := getAutoReviewers(repo, conf)
		if err != nil {
			logger.Errorf("failed to init reviewer for repo: %s/%s with err: %v", repo.ProjectName, repo.Name, err)
			continue
		}

		aReviewers = append(aReviewers, aReviewer)
	}

	for _, aReviewer := range aReviewers {
		logger.Infof("Starting Reviewer for repo: %s\n", aReviewer.Repository)
		if err := aReviewer.Run(); err != nil {
			logger.Errorf("Failed to balance repo: %s with err: %v\n", aReviewer.Repository, err)
		}
		logger.Infof("Finished Balancing Cycle for repo: %s\n", aReviewer.Repository)
	}

	logger.Info("Finished Reviewing for all repositories")
	return nil
}

func getAutoReviewers(repo *types.Repository, conf *config.Config) (*autoreviewer.AutoReviewer, error) {
	vstsConfig := &vsts.Config{
		Token:          conf.Token,
		Username:       conf.Username,
		APIVersion:     conf.APIVersion,
		RepositoryName: repo.Name,
		Project:        repo.ProjectName,
		Instance:       conf.Instance,
	}

	fmt.Printf("VSTSCONFIG: %+v\n", vstsConfig)
	fmt.Println("############")
	fmt.Printf("REPO: %+v\n", repo)

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
			log.G(context.TODO()).Errorf("ERROR: Failed to create slack trigger with error: %v", err)
		} else {
			reviewerTriggers = append(reviewerTriggers, slackTrigger)
			log.G(context.TODO()).Info("Adding Slack Reviewer Trigger...")
		}
	}

	aReviewer, err := autoreviewer.NewAutoReviewer(vstsClient, conf.BotMaker, repo.ReviewerGroups, filters, reviewerTriggers)
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
