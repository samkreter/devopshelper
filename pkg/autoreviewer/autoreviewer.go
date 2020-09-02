package autoreviewer

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/samkreter/devopshelper/pkg/utils"
	"strings"
	"time"

	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	adoidentity "github.com/microsoft/azure-devops-go-api/azuredevops/identity"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/samkreter/go-core/log"

	"github.com/samkreter/devopshelper/pkg/store"
	"github.com/samkreter/devopshelper/pkg/types"
)

const (
	defaultBotIdentifier = "b03f5f7f11d50a3a"
)

var (
	defaultFilters = []Filter{
		filterWIP,
		filterMasterBranchOnly,
		filterBotV2PRs,
	}
)

// Filter is a function returns true if a pull request should be filtered out.
type Filter func(*PullRequest) bool

// ReviewerTrigger is called with the reviewers that have been selected. Allows for adding custom events
//  for each reviewer that is added to the PR. Ex: slack notification.
type ReviewerTrigger func([]*types.Reviewer, []*types.Reviewer, string) error

type Options struct {
	Filters []Filter
	ReviewerTriggers []ReviewerTrigger
}

// AutoReviewer automaticly adds reviewers to a vsts pull request
type AutoReviewer struct {
	adoGitClient     adogit.Client
	adoIdentityClient adoidentity.Client
	adoCoreClient adocore.Client
	botIdentifier         string
	Repo             *types.Repository
	RepoStore        store.RepositoryStore
	Options Options
}

// NewAutoReviewer creates a new autoreviewer
func NewAutoReviewer(adoGitClient adogit.Client,
	adoIdentityClient adoidentity.Client, adoCoreClient adocore.Client,
	botIdentifier string, repo *types.Repository, repoStore store.RepositoryStore,
	options Options) (*AutoReviewer, error) {

	if options.Filters == nil {
		options.Filters = defaultFilters
	}

	return &AutoReviewer{
		Repo:              repo,
		RepoStore:         repoStore,
		Options:           options,
		adoGitClient:      adoGitClient,
		adoIdentityClient: adoIdentityClient,
		adoCoreClient:     adoCoreClient,
		botIdentifier:     botIdentifier,
	}, nil
}

// Run starts the autoreviewer for a single instance
func (a *AutoReviewer) Run(ctx context.Context) error {
	pullRequests, err := a.adoGitClient.GetPullRequests(ctx, adogit.GetPullRequestsArgs{
		RepositoryId: &a.Repo.AdoRepoID,
		Project: &a.Repo.ProjectName,
		SearchCriteria: &adogit.GitPullRequestSearchCriteria{},
	})
	if err != nil {
		return fmt.Errorf("get pull requests error: %v", err)
	}



	for _, pr := range *pullRequests {
		pullRequest := &PullRequest{pr}

		if a.shouldFilter(pullRequest) {
			continue
		}

		if err := a.balanceReview(ctx, pullRequest); err != nil {
			return errors.Wrap(err, "failed to balancer reviewers")
		}
	}

	return nil
}

func (a *AutoReviewer) Reconcile(ctx context.Context) error {
	if err := a.ensureAdoRepoID(ctx); err != nil {
		return errors.Wrap(err, "failed to ensure repo id")
	}

	if err := a.ensureReviewers(ctx); err != nil {
		return errors.Wrap(err, "failed to ensure reviewers")
	}

	return nil
}

func (a *AutoReviewer) balanceReview(ctx context.Context, pr *PullRequest) error {
	logger := log.G(ctx)

	if a.ContainsReviewBalancerComment(ctx, pr.Repository.Id.String(),  *pr.PullRequestId) {
		return nil
	}

	requiredReviewers, optionalReviewers, err := a.getReviewers(ctx, pr)
	if err != nil {
		return errors.Wrap(err, "failed to get reviewers")
	}

	if err := a.AddReviewers(ctx, *pr.PullRequestId, pr.Repository.Id.String(), requiredReviewers, optionalReviewers); err != nil {
		return errors.Wrap(err, "failed to add reviewers to PR")
	}

	if err := a.addReviewerComment(ctx, pr, requiredReviewers); err != nil {
		return errors.Wrap(err,"failed to add reviewer comment")
	}

	if a.Options.ReviewerTriggers != nil {
		for _, rTrigger := range a.Options.ReviewerTriggers {
			if err := rTrigger(requiredReviewers, optionalReviewers, *pr.Url); err != nil {
				logger.Error(err)
			}
		}
	}

	logger.Infof("Successfully added %s as required reviewers and %s as observer to PR: %d",
		GetReviewersAlias(requiredReviewers),
		GetReviewersAlias(optionalReviewers),
		*pr.PullRequestId)

	return nil
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

		if reviewerGroup.Team != "" {
			members, err := a.getTeamMembers(ctx, reviewerGroup.Team)
			if err != nil {
				return errors.Wrap(err, "failed to get team memebers")
			}

			for _, member := range members {
				reviewerAliases[member] = true
			}
		}
	}

	// ensure reviewers are up to date in the DB
	for alias := range reviewerAliases {
		reviewer, err := a.RepoStore.GetReviewer(ctx, alias)
		if err != nil {
			switch {
			// Add the reviewer if it doens't exist
			case errors.Is(err, store.ErrNotFound):
				reviewer, err := utils.GetReviewerFromAlias(ctx, alias, a.adoIdentityClient)
				if err != nil {
					return errors.Wrap(err, "failed to get reviewer from alias")
				}
				a.RepoStore.AddReviewer(ctx, reviewer)
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
			a.RepoStore.UpdateReviewer(ctx, reviewer)
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

func (a *AutoReviewer) shouldFilter(pr *PullRequest) bool {
	if a.Options.Filters == nil {
		return false
	}

	for _, filter := range a.Options.Filters {
		if filter(pr) {
			return true
		}
	}

	return false
}

func (a *AutoReviewer) getReviewers(ctx context.Context, pr *PullRequest) ([]*types.Reviewer, []*types.Reviewer, error) {
	reviewerGroups, err := pr.GetRequiredReviewerGroups(ctx, a.adoGitClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get required reviewer groups")
	}

	requiredOwners := map[string]bool{}
	requiredTeamMembers := map[string]bool{}

	for _, reviewerGroup := range reviewerGroups {
		if reviewerGroup == nil {
			continue
		}

		if reviewerGroup.Team != "" {
			teamMembers, err := a.getTeamMembers(ctx, reviewerGroup.Team)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "failed to get team members for team %q", reviewerGroup.Team)
			}

			for _, member := range teamMembers {
				requiredTeamMembers[member] = true
			}
		}

		for owner := range reviewerGroup.Owners {
			requiredOwners[owner] = true
		}
	}

	// Ensure owners aren't in both groups
	for owner := range requiredOwners {
		delete(requiredTeamMembers, owner)
	}

	// Get least recently used reviewer for each group
	owners := getAliases(requiredOwners)
	owner, err := a.RepoStore.PopLRUReviewer(ctx, owners)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get owner reviewer")
	}

	teamMembers := getAliases(requiredTeamMembers)
	teamMember, err := a.RepoStore.PopLRUReviewer(ctx, teamMembers)
	if err != nil {
		switch {
			case errors.Is(err, store.ErrNotFound):
				return []*types.Reviewer{owner}, nil, nil
		default:
			return nil, nil, errors.Wrapf(err, "failed to get team reviewer for members: %v", teamMembers)
		}
	}

	return []*types.Reviewer{owner, teamMember}, nil, nil
}

func getAliases(reviewers map[string]bool) []string {
	if reviewers == nil {
		return nil
	}

	aliases := make([]string, 0, len(reviewers))
	for alias, enabled := range reviewers {
		if enabled {
			aliases = append(aliases, alias)
		}
	}

	return aliases
}

func (a *AutoReviewer) addReviewerComment(ctx context.Context, pr *PullRequest, required []*types.Reviewer) error {
	comment := fmt.Sprintf(
		"Hello %s,\r\n\r\n"+
			"You are randomly selected as the **required** code reviewers of this change. \r\n\r\n"+
			"Your responsibility is to review **each** iteration of this CR until signoff. You should provide no more than 48 hour SLA for each iteration.\r\n\r\n"+
			"Thank you.\r\n\r\n"+
			"CR Balancer\r\n"+
			"%s",
		strings.Join(GetReviewersAlias(required), ","),
		a.botIdentifier)

	repoID := pr.Repository.Id.String()
	_, err := a.adoGitClient.CreateThread(ctx, adogit.CreateThreadArgs{
		RepositoryId: &repoID,
		PullRequestId: pr.PullRequestId,
		CommentThread: &adogit.GitPullRequestCommentThread{
			Comments: &[]adogit.Comment{
				{
					Content: &comment,
				},
			},
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to add reviewer comment")
	}

	return nil
}

// ContainsReviewBalancerComment checks if the passed in review has had a bot comment added.
func (a *AutoReviewer) ContainsReviewBalancerComment(ctx context.Context, repositoryID string, pullRequestID int) bool {
	threads, err := a.adoGitClient.GetThreads(ctx, adogit.GetThreadsArgs{
		RepositoryId: &repositoryID,
		PullRequestId: &pullRequestID,
	})
	if err != nil {
		panic(err)
	}

	if threads != nil {
		for _, thread := range *threads {
			for _, comment := range *thread.Comments {
				if comment.Content == nil {
					continue
				}
				if strings.Contains(*comment.Content, a.botIdentifier) {
					return true
				}
			}
		}
	}
	return false
}

// AddReviewers adds the passing in reviewers to the pull requests for the passed in review.
func (a *AutoReviewer) AddReviewers(ctx context.Context, pullRequestID int, repoID string, required, optional []*types.Reviewer) error {
	for _, reviewer := range required {
		_, err := a.adoGitClient.CreatePullRequestReviewer(ctx, adogit.CreatePullRequestReviewerArgs{
			Reviewer: &adogit.IdentityRefWithVote{
				IsRequired: toBoolPtr(true),
			},
			ReviewerId: &reviewer.AdoID,
			RepositoryId: &repoID,
			PullRequestId: &pullRequestID,
		})
		if err != nil {
			return fmt.Errorf("failed to add required reviewer %s with error %w", reviewer.Alias, err)
		}
	}

	for _, reviewer := range optional {
		_, err := a.adoGitClient.CreatePullRequestReviewer(ctx, adogit.CreatePullRequestReviewerArgs{
			Reviewer: &adogit.IdentityRefWithVote{
				IsRequired: toBoolPtr(false),
			},
			ReviewerId: &reviewer.AdoID,
			RepositoryId: &repoID,
			PullRequestId: &pullRequestID,
		})
		if err != nil {
			return fmt.Errorf("failed to add optional reviewer %s with error %w", reviewer.Alias, err)
		}
	}

	return nil
}

// GetReviewersAlias gets all names for the set of passed in reviewers
// return: string slice of the aliases
func GetReviewersAlias(reviewers []*types.Reviewer) []string {
	aliases := make([]string, len(reviewers))

	for index, reviewer := range reviewers {
		aliases[index] = reviewer.Alias
	}
	return aliases
}

func filterWIP(pr *PullRequest) bool {
	if strings.Contains(*pr.Title, "WIP") {
		return true
	}

	return false
}

func filterBotV2PRs(pr *PullRequest) bool {
	if !strings.Contains(*pr.Title, "BOTv2") {
		return true
	}

	return false
}

func filterMasterBranchOnly(pr *PullRequest) bool {
	if strings.EqualFold(*pr.TargetRefName, "refs/heads/master") {
		return false
	}

	return true
}
