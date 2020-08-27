package main

import (
	"context"
	"flag"
	"fmt"
	adoidentity "github.com/microsoft/azure-devops-go-api/azuredevops/identity"
	"strings"
	"time"

	"github.com/samkreter/go-core/log"
	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops"

	"github.com/samkreter/devopshelper/pkg/autoreviewer"
	"github.com/samkreter/devopshelper/pkg/server"
	"github.com/samkreter/devopshelper/pkg/store"
)

const (
	defaultReviewerIntervalMin = 5
)

var (
	adminsStr      string
	enableCORS     bool
	adoPatToken      string
	botIdentifier string
	organizationUrl string
	logLvl         string
	mongoOptions   = &store.MongoStoreOptions{}
	serverOptions  = &server.Options{}
)

func main() {
	flag.StringVar(&serverOptions.Addr, "addr", "localhost:8080", "the address for the api server to listen on.")
	flag.StringVar(&adminsStr, "admins", "", "admins to be added to each repo, comma seperated")
	flag.BoolVar(&serverOptions.AllowCORS, "enable-cors", true, "enable cors for the api server.")

	flag.StringVar(&logLvl, "log-level", "info", "the log level for the application")

	flag.StringVar(&adoPatToken, "pat-token", "", "vsts personal access token")
	flag.StringVar(&botIdentifier, "botmaker-id", "b03f5f7f11d50a3a", "identifier for the bot's message")
	flag.StringVar(&organizationUrl, "organizationUrl", "https://msazure.visualstudio.com", "vsts instance")

	flag.StringVar(&mongoOptions.MongoURI, "mongo-uri", "", "connection string for the mongo database")
	flag.StringVar(&mongoOptions.RepositoryCollection, "mongo-repo-collection", "", "collection that stores the repositories")
	flag.StringVar(&mongoOptions.BaseGroupCollection, "mongo-basegroup-collection", "", "collection that stores the base groups")
	flag.StringVar(&mongoOptions.DBName, "mongo-dbname", "reviewerBot", "the mongo database to access")
	flag.BoolVar(&mongoOptions.UseSSL, "mongo-ssl", false, "use ssl when accessing mongo database")

	reviewIntervalMin := flag.Int("review-interval", defaultReviewerIntervalMin, "number of minutes to wait to reviwer")
	flag.Parse()

	ctx := context.Background()
	logger := log.G(ctx)

	if err := log.SetLogLevel(logLvl); err != nil {
		logger.Errorf("failed to set log level to : '%s'", logLvl)
	}
	
	serverOptions.Admins = strings.Split(adminsStr, ",")

	repoStore, err := store.NewMongoStore(mongoOptions)
	if err != nil {
		logger.Fatal(err)
	}
	r, err := repoStore.PopLRUReviewer(ctx, []string{"segoings", "segoings2"})
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Println(r)

	return

	conn := azuredevops.NewPatConnection(organizationUrl, adoPatToken)
	adoGitClient, err := adogit.NewClient(ctx, conn)
	if err != nil {
		logger.Fatal(err)
	}

	adoIdentityClient, err := adoidentity.NewClient(ctx, conn)
	if err != nil {
		logger.Fatal(err)
	}

	go func() {
		logger.Info("Starting Reviewer Reconcile Loop....")

		mgr, err := autoreviewer.NewDefaultManager(ctx, repoStore, adoGitClient, adoIdentityClient)
		if err != nil {
			logger.Errorf("Failed to create reviewer manager: %s", err)
		}
		mgr.Run(ctx)
		logger.Info("Finished Reviewing for all repositories")

		for {
			select {
			case <-time.NewTicker(time.Minute * time.Duration(*reviewIntervalMin)).C:
				mgr, err := autoreviewer.NewDefaultManager(ctx, repoStore, adoGitClient, adoIdentityClient)
				if err != nil {
					logger.Errorf("Failed to create reviewer manager: %s", err)
					continue
				}
				mgr.Run(ctx)
				logger.Info("Finished Reviewing for all repositories")

			case <-ctx.Done():
				logger.Info(fmt.Sprintf("shutting down reviewer loop: %s", ctx.Err()))
				return
			}
		}
	}()

	s, err := server.NewServer(adoGitClient, adoIdentityClient, repoStore, serverOptions)
	if err != nil {
		logger.Fatal(err)
	}

	s.Run()
}


