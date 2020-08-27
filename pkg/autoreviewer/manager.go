package autoreviewer

import (
	"context"

	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	adoidentity "github.com/microsoft/azure-devops-go-api/azuredevops/identity"
	"github.com/samkreter/go-core/log"

	"github.com/samkreter/devopshelper/pkg/store"
	"github.com/samkreter/devopshelper/pkg/types"
)

type Manager struct {
	AutoReviewers []*AutoReviewer
	repoStore store.RepositoryStore
}

func NewDefaultManager(ctx context.Context, repoStore store.RepositoryStore, adoGitClient adogit.Client, aodIdentityClient adoidentity.Client) (*Manager, error) {
	repos, err := repoStore.GetAllRepositories(ctx)
	if err != nil {
		return nil, err
	}

	enabledRepos := []*types.Repository{}
	for _, repo := range repos {
		if repo.Enabled {
			enabledRepos = append(enabledRepos, repo)
		}
	}

	aReviewers := make([]*AutoReviewer, 0, len(repos))
	for _, repo := range enabledRepos {
		aReviewer, err := NewAutoReviewer(adoGitClient, aodIdentityClient, defaultBotIdentifier, repo, repoStore, defaultFilters, nil)
		if err != nil {
			return nil, err
		}

		aReviewers = append(aReviewers, aReviewer)
	}

	return &Manager{
		repoStore: repoStore,
		AutoReviewers: aReviewers,
	}, nil
}

func (m *Manager) Run(ctx context.Context) {
	logger := log.G(ctx)

	for _, aReviewer := range m.AutoReviewers {
		logger.Infof("Starting Reviewer for repo: %s/%s", aReviewer.Repo.ProjectName, aReviewer.Repo.Name)
		if err := aReviewer.Run(ctx); err != nil {
			logger.Errorf("Failed to balance repo: %s/%s with err: %v", aReviewer.Repo.ProjectName, aReviewer.Repo.Name, err)
		}
		logger.Infof("Finished Balancing Cycle for: %s/%s", aReviewer.Repo.ProjectName, aReviewer.Repo.Name)
	}
}
