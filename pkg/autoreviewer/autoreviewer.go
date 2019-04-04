package autoreviewer

import (
	"fmt"
	"log"
	"strings"
	"time"

	vstsObj "github.com/samkreter/vsts-goclient/api/git"
	vsts "github.com/samkreter/vsts-goclient/client"
	"github.com/samkreter/vstsautoreviewer/pkg/types"
)

// Filter is a function returns true if a pull request should be filtered out.
type Filter func(vstsObj.GitPullRequest) bool

// ReviwerTrigger is called with the reviewers that have been selected. Allows for adding custom events
//  for each reviewer that is added to the PR. Ex: slack notification.
type ReviwerTrigger func([]*types.Reviewer, string) error

// AutoReviewer automaticly adds reviewers to a vsts pull request
type AutoReviewer struct {
	Repository       string
	filters          []Filter
	reviewerTriggers []ReviwerTrigger
	vstsClient       *vsts.Client
	reviewerGroups   types.ReviewerGroups
	reviewerFile     string
	statusFile       string
	botMaker         string
}

// ReviewerInfo describes who to be added as a reviwer and which files to watch for
type ReviewerInfo struct {
	File           string   `json:"file"`
	ActivePaths    []string `json:"activePaths"`
	reviewerGroups types.ReviewerGroups
}

// NewAutoReviewer creates a new autoreviewer
func NewAutoReviewer(vstsClient *vsts.Client, botMaker string, reviewerGroups types.ReviewerGroups, filters []Filter, rTriggers []ReviwerTrigger) (*AutoReviewer, error) {
	return &AutoReviewer{
		Repository:       vstsClient.Repo,
		vstsClient:       vstsClient,
		filters:          filters,
		reviewerTriggers: rTriggers,
		reviewerGroups:   reviewerGroups,
		botMaker:         botMaker,
	}, nil
}

// RunInterval executes the autoreviewer at the set interval
func (a *AutoReviewer) RunInterval() {
	for range time.NewTicker(time.Second * 30).C {
		if err := a.Run(); err != nil {
			log.Println("Error for main run", err)
		}
		log.Println("Running Review cycle...")
	}
}

func (a *AutoReviewer) shouldFilter(pr vstsObj.GitPullRequest) bool {
	for _, filter := range a.filters {
		if filter(pr) {
			return true
		}
	}

	return false
}

// Run starts the autoreviewer for a single instance
func (a *AutoReviewer) Run() error {
	pullRequests, err := a.vstsClient.GetPullRequests(nil)
	if err != nil {
		return fmt.Errorf("get pull requests error: %v", err)
	}

	for _, pullRequest := range pullRequests {
		if a.shouldFilter(pullRequest) {
			continue
		}

		if err := a.balanceReview(pullRequest); err != nil {
			log.Printf("ERROR: balancing reviewers with error %v", err)
		}
	}

	return nil
}

func (a *AutoReviewer) balanceReview(pullRequest vstsObj.GitPullRequest) error {
	if !a.ContainsReviewBalancerComment(pullRequest.PullRequestId) {
		requiredReviewers, optionalReviewers, err := a.reviewerGroups.GetReviewers(pullRequest.CreatedBy.ID, a.statusFile)
		if err != nil {
			return err
		}

		if err := a.AddReviewers(pullRequest.PullRequestId, requiredReviewers, optionalReviewers); err != nil {
			return fmt.Errorf("add reviewers error: %v", err)
		}

		comment := fmt.Sprintf(
			"Hello %s,\r\n\r\n"+
				"You are randomly selected as the **required** code reviewers of this change. \r\n\r\n"+
				"Your responsibility is to review **each** iteration of this CR until signoff. You should provide no more than 48 hour SLA for each iteration.\r\n\r\n"+
				"Thank you.\r\n\r\n"+
				"CR Balancer\r\n"+
				"%s",
			strings.Join(GetReviewersAlias(requiredReviewers), ","),
			a.botMaker)

		if err := a.vstsClient.AddThread(pullRequest.PullRequestId, comment); err != nil {
			return fmt.Errorf("add thread error: %v", err)
		}
		log.Printf("INFO: Adding %s as required reviewers and %s as observer to PR: %d",
			GetReviewersAlias(requiredReviewers),
			GetReviewersAlias(requiredReviewers),
			pullRequest.PullRequestId)

		pullRequestURL := getPullRequestURL(a.vstsClient.Instance, a.vstsClient.Project, a.vstsClient.Repo, pullRequest.PullRequestId)
		for _, rTrigger := range a.reviewerTriggers {
			if err := rTrigger(requiredReviewers, pullRequestURL); err != nil {
				log.Printf("ERROR: %v", err)
			}
		}
	}

	return nil
}

// ContainsReviewBalancerComment checks if the passed in review has had a bot comment added.
func (a *AutoReviewer) ContainsReviewBalancerComment(pullRequestID int32) bool {
	threads, err := a.vstsClient.GetThreads(pullRequestID, nil)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	if threads != nil {
		for _, thread := range threads {
			for _, comment := range thread.Comments {
				if strings.Contains(comment.Content, a.botMaker) {
					return true
				}
			}
		}
	}
	return false
}

// AddReviewers adds the passing in reviewers to the pull requests for the passed in review.
func (a *AutoReviewer) AddReviewers(pullRequestID int32, required, optional []*types.Reviewer) error {
	for _, reviewer := range append(required, optional...) {
		identity := vstsObj.IdentityRefWithVote{
			ID: reviewer.ID,
		}

		if err := a.vstsClient.AddReviewer(pullRequestID, identity); err != nil {
			log.Printf("WARN: Failed to add reviewer %s with error %v", reviewer.Alias, err)
			continue
		}
	}

	return nil
}

func getPullRequestURL(instance, project, repository string, pullRequestID int32) string {
	return fmt.Sprintf(
		"https://%s/DefaultCollection/%s/_git/%s/pullrequest/%d",
		instance,
		project,
		repository,
		pullRequestID)
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
