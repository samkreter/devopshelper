package autoreviewer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/samkreter/go-core/log"

	"github.com/samkreter/devopshelper/pkg/store"
	"github.com/samkreter/devopshelper/pkg/utils"
)

func (a *AutoReviewer) Reconcile(ctx context.Context) error {
	if err := a.ensureAdoRepoID(ctx); err != nil {
		return errors.Wrap(err, "failed to ensure repo id")
	}

	if err := a.ensureReviewers(ctx); err != nil {
		return errors.Wrap(err, "failed to ensure reviewers")
	}

	return nil
}

func (a *AutoReviewer) ensureAdoRepoID(ctx context.Context) error{
	logger := log.G(ctx)
	logger.Infof("Starting Repo ID reconciling for repo: %s", a.Repo.Name)

	if a.Repo.AdoRepoID != "" {
		return nil
	}

	adoRepos, err := a.adoGitClient.GetRepositories(ctx, adogit.GetRepositoriesArgs{
		Project: &a.Repo.ProjectName,
	})
	if err != nil {
		return err
	}

	for _, adoRepo := range *adoRepos {
		if *adoRepo.Name == a.Repo.Name {
			a.Repo.AdoRepoID = adoRepo.Id.String()
			if err := a.RepoStore.UpdateRepository(ctx, a.Repo.ID.Hex(), a.Repo); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("repo: %s not found in project %s", a.Repo.Name, a.Repo.ProjectName)
}

func (a *AutoReviewer) ensureReviewers(ctx context.Context) error {
	logger := log.G(ctx)
	logger.Infof("Starting repo reviewers reconciling for repo: %s", a.Repo.Name)

	items, err := a.adoGitClient.GetItems(ctx, adogit.GetItemsArgs{
		RepositoryId:   &a.Repo.AdoRepoID,
		RecursionLevel: &adogit.VersionControlRecursionTypeValues.Full,
	})
	if err != nil {
		return errors.Wrap(err, "failed to get ownersfiles")
	}

	// Get all reviewer groups for the repo
	reviewerAliases := map[string]bool{}
	for _, item := range *items {
		if !strings.Contains(*item.Path, "owners.txt") {
			continue
		}

		ownersFile, err := getFileFromADO(ctx, a.adoGitClient, a.Repo.AdoRepoID, *item.Path)
		if err != nil {
			return err
		}

		reviewerGroup :=  newReviewerGroupFromOwnersFile(*ownersFile.Content)

		for owner := range reviewerGroup.Owners {
			reviewerAliases[owner] = true
		}

		for team := range reviewerGroup.Teams {
			members, err := a.getTeamMembers(ctx, team)
			if err != nil {
				return errors.Wrapf(err, "failed to get team members for team: %s", team)
			}

			for _, member := range members {
				reviewerAliases[member] = true
			}
		}
	}

	// ensure reviewers are up to date in the DB
	for alias := range reviewerAliases {
		reviewer, err := a.ReviewerStore.GetReviewer(ctx, alias)
		if err != nil {
			switch {
			// Add the reviewer if it doens't exist
			case errors.Is(err, store.ErrNotFound):
				reviewer, err := utils.GetReviewerFromAlias(ctx, alias, a.adoIdentityClient)
				if err != nil {
					return errors.Wrap(err, "failed to get reviewer from alias")
				}
				a.ReviewerStore.AddReviewer(ctx, reviewer)
				logger.Infof("Adding new reviwer: %s", alias)
				continue
			default:
				return errors.Wrap(err, "failed to get reviewer from store")
			}
		}

		// ensure we have the ADO ID
		if reviewer.AdoID == "" {
			reviewer, err := utils.GetReviewerFromAlias(ctx, alias, a.adoIdentityClient)
			if err != nil {
				return err
			}
			a.ReviewerStore.UpdateReviewer(ctx, reviewer)
			logger.Infof("Updating reviewer: %q with ado id", alias)
			continue
		}
	}

	a.Repo.LastReconciled = time.Now().UTC()
	if err := a.RepoStore.UpdateRepository(ctx,a.Repo.ID.Hex(), a.Repo); err != nil {
		return err
	}

	return nil
}
