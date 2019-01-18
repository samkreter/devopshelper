package autoreviewer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	vstsObj "github.com/samkreter/vsts-goclient/api/git"
	vsts "github.com/samkreter/vsts-goclient/client"
)

// Filter is a function returns true if a pull request should be filtered out.
type Filter func(vstsObj.GitPullRequest) bool

// AutoReviewer automaticly adds reviewers to a vsts pull request
type AutoReviewer struct {
	Repository     string
	filters        []Filter
	vstsClient     *vsts.Client
	reviewerGroups *ReviewerGroups
	reviewerFile   string
	statusFile     string
	botMaker       string
}

// NewAutoReviewer creates a new autoreviewer
func NewAutoReviewer(vstsClient *vsts.Client, botMaker, reviewerFile, statusFile string, filters []Filter) (*AutoReviewer, error) {
	reviewerGroups, err := loadReviewerGroups(reviewerFile, statusFile)
	if err != nil {
		return nil, err
	}

	return &AutoReviewer{
		Repository:     vstsClient.Repo,
		vstsClient:     vstsClient,
		filters:        filters,
		reviewerGroups: reviewerGroups,
		reviewerFile:   reviewerFile,
		statusFile:     statusFile,
		botMaker:       botMaker,
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
		requiredReviewers, optionalReviewers := a.reviewerGroups.GetReviewers(pullRequest.CreatedBy.ID, a.statusFile)

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
func (a *AutoReviewer) AddReviewers(pullRequestID int32, required, optional []Reviewer) error {
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

// ReviewerGroups is a list of type ReviewerGroup
type ReviewerGroups []ReviewerGroup

// ReviewerPositions holds the current position information for the reviewers
type ReviewerPositions map[string]int

// ReviewerGroup holds the reviwers and metadata for a review group.
type ReviewerGroup struct {
	Group      string     `json:"group"`
	Required   bool       `json:"required"`
	Reviewers  []Reviewer `json:"reviewers"`
	CurrentPos int
}

// Reviewer is a vsts revier object
type Reviewer struct {
	UniqueName string `json:"uniqueName"`
	Alias      string `json:"alias"`
	ID         string `json:"id"`
}

func loadReviewerGroups(reviewerFile, statusFile string) (*ReviewerGroups, error) {
	rawReviewerData, err := ioutil.ReadFile(reviewerFile)
	if err != nil {
		log.Fatal("Could not load ", reviewerFile)
	}

	var reviewerGroups ReviewerGroups
	json.Unmarshal(rawReviewerData, &reviewerGroups)

	rawPosData, err := ioutil.ReadFile(statusFile)
	if err != nil {
		return nil, fmt.Errorf("could not load current state file %s", statusFile)
	}

	var reviewerPoses ReviewerPositions
	json.Unmarshal(rawPosData, &reviewerPoses)

	for index, reviewerGroup := range reviewerGroups {
		if pos, ok := reviewerPoses[reviewerGroup.Group]; ok {
			reviewerGroups[index].CurrentPos = pos
		}
	}

	return &reviewerGroups, nil
}

func (rg *ReviewerGroups) savePositions(statusFile string) {
	reviewerPositions := make(ReviewerPositions)
	for _, reviewerGroup := range *rg {
		reviewerPositions[reviewerGroup.Group] = reviewerGroup.CurrentPos
	}

	data, err := json.Marshal(reviewerPositions)
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(statusFile, data, 0644); err != nil {
		log.Fatal(err)
	}

	log.Println("INFO: Saving position file.")
}

func (g *ReviewerGroup) getCurrentReviewer() Reviewer {
	return g.Reviewers[g.CurrentPos]
}

func (g *ReviewerGroup) incPos() {
	g.CurrentPos = (g.CurrentPos + 1) % len(g.Reviewers)
}

// GetReviewersAlias gets all names for the set of passed in reviewers
// return: string slice of the aliases
func GetReviewersAlias(reviewers []Reviewer) []string {
	aliases := make([]string, len(reviewers))

	for index, reviewer := range reviewers {
		aliases[index] = reviewer.Alias
	}
	return aliases
}

// GetReviewers gets the required and optional reviewers for a review
// review: the review summary
// return: returns a slice of require reviewers and a slice of optional reviewers
func (rg *ReviewerGroups) GetReviewers(pullRequestCreatorID, statusFile string) ([]Reviewer, []Reviewer) {
	requiredReviewers := make([]Reviewer, 0, len(*rg)/2)
	optionalReviewers := make([]Reviewer, 0, len(*rg)/2)

	for index := range *rg {
		if (*rg)[index].Required == true {
			requiredReviewers = append(requiredReviewers, getNextReviewer(&(*rg)[index], pullRequestCreatorID))
		} else {
			optionalReviewers = append(optionalReviewers, getNextReviewer(&(*rg)[index], pullRequestCreatorID))
		}
	}

	rg.savePositions(statusFile)
	return requiredReviewers, optionalReviewers
}

func getNextReviewer(group *ReviewerGroup, pullRequestCreatorID string) Reviewer {
	defer group.incPos()

	for len(group.Reviewers) > 1 && group.getCurrentReviewer().ID == pullRequestCreatorID {

		group.incPos()
	}

	return group.getCurrentReviewer()
}
