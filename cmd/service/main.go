package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/samkreter/go-core/log"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/core"
	adoidentity "github.com/microsoft/azure-devops-go-api/azuredevops/identity"

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

	mongoStore, err := store.NewMongoStore(mongoOptions)
	if err != nil {
		logger.Fatal(AddStack(err))

	}

	conn := azuredevops.NewPatConnection(organizationUrl, adoPatToken)
	adoGitClient, err := adogit.NewClient(ctx, conn)
	if err != nil {
		logger.Fatal(AddStack(err))
	}

	adoIdentityClient, err := adoidentity.NewClient(ctx, conn)
	if err != nil {
		logger.Fatal(AddStack(err))
	}

	adoCoreClient, err := adocore.NewClient(ctx, conn)
	if err != nil {
		logger.Fatal(AddStack(err))
	}

	// TODO: Handle errors
	go func() {
		logger.Info("Starting Reviewer Reconcile Loop....")

		mgr, err := autoreviewer.NewDefaultManager(ctx, mongoStore, mongoStore, adoGitClient, adoIdentityClient, adoCoreClient)
		if err != nil {
			logger.Errorf("Failed to create reviewer manager: %s", err)
			return
		}
		if err := mgr.Run(ctx); err != nil {
			logger.Fatal(AddStack(err))
		}
		logger.Info("Finished Reviewing for all repositories")

		for {
			select {
			case <-time.NewTicker(time.Minute * time.Duration(*reviewIntervalMin)).C:
				mgr, err := autoreviewer.NewDefaultManager(ctx, mongoStore, mongoStore, adoGitClient, adoIdentityClient, adoCoreClient)
				if err != nil {
					logger.Errorf("Failed to create reviewer manager: %s", err)
					continue
				}
				if err := mgr.Run(ctx); err != nil {
					logger.Fatal(AddStack(err))
				}
				logger.Info("Finished Reviewing for all repositories")

			case <-ctx.Done():
				logger.Info(fmt.Sprintf("shutting down reviewer loop: %s", ctx.Err()))
				return
			}
		}
	}()

	s, err := server.NewServer(adoGitClient, adoIdentityClient, mongoStore, serverOptions)
	if err != nil {
		logger.Fatal(AddStack(err))
	}

	s.Run()
}

func AddStack(err error) error {
	stack := getStackTrace(err)
	if stack == "" {
		return err
	}

	return fmt.Errorf("%w\n%s", err, stack)
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func getStackTrace(err error) string {
	stErr := getBaseStackTracer(err)
	if stErr != nil {
		st := stErr.StackTrace()
		return fmt.Sprintf("%+v", st)
	}

	return ""
}

func getBaseStackTracer(err error) stackTracer {
	type unwrapper interface {
		Unwrap() error
	}

	var baseST stackTracer

	for err != nil {
		if st, ok := err.(stackTracer); ok {
			baseST = st
		}

		unwrappedErr, ok := err.(unwrapper)
		if !ok {
			break
		}
		err = unwrappedErr.Unwrap()
	}
	return baseST
}


