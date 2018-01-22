package vsts

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var (
	ReviewerFile = "./configs/reviewers.json"
	StatusFile = "./configs/currentStatus.json"

	reviewerGroups ReviewerGroups
)

// ReviewerGroups is a list of type ReviewerGroup
type ReviewerGroups []ReviewerGroup

func (reviewerGroups *ReviewerGroups) loadReviewerGroups() {
	rawReviewerData, err := ioutil.ReadFile(ReviewerFile)
	if err != nil {
		log.Fatal("Could not load ", ReviewerFile)
	}

	json.Unmarshal(rawReviewerData, reviewerGroups)

	rawPosData, err := ioutil.ReadFile(StatusFile)
	if err != nil {
		log.Printf("Could not load %s. Using default positions.", StatusFile)
		return
	}
	
	var reviewerPoses ReviewerPositions
	json.Unmarshal(rawPosData, &reviewerPoses)

	for _, reviewerGroup := range *reviewerGroups{
		if pos, ok := reviewerPoses[reviewerGroup.Group]; ok{
			reviewerGroup.CurrentPos = pos
		}
	}
}

func (reviewerGroups ReviewerGroups) savePositions(){
	reviewerPositions := make(ReviewerPositions)
	for _,  reviewerGroup := range reviewerGroups {
		reviewerPositions[reviewerGroup.Group] = reviewerGroup.CurrentPos
	}
	
	data, err := json.Marshal(reviewerPositions)
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(StatusFile, data, 0644); err != nil {
		log.Fatal(err)
	}
}

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
	VisualStudioID string `json:"id"`
	Email          string `json:"uniqueName"`
	Alias          string `json:"alias"`
}

func (g ReviewerGroup) getCurrentReviewer() Reviewer {
	return g.Reviewers[g.CurrentPos]
}

func (g *ReviewerGroup) incPos() {
	g.CurrentPos = (g.CurrentPos + 1) % len(g.Reviewers)
}

// ReviewSummary holds information for a review.
type ReviewSummary struct {
	ID           string
	AuthorAlias  string
	AuthorEmail  string
	AuthorVstsID string
	RepositoryID string
	ReviewType   string
}

func init() {
	reviewerGroups.loadReviewerGroups()
}

// GetReviewersAlias gets all aliases for the set of passed in reviewers
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
func GetReviewers(review ReviewSummary) ([]Reviewer, []Reviewer) {
	requiredReviewers := make([]Reviewer, 0, len(reviewerGroups)/2)
	optionalReviewers := make([]Reviewer, 0, len(reviewerGroups)/2)

	for index := range reviewerGroups {
		if reviewerGroups[index].Required == true {
			requiredReviewers = append(requiredReviewers, getNextReviewer(&reviewerGroups[index], review))
		} else {
			optionalReviewers = append(optionalReviewers, getNextReviewer(&reviewerGroups[index], review))
		}
	}
	
	reviewerGroups.savePositions()
	return requiredReviewers, optionalReviewers
}

func getNextReviewer(group *ReviewerGroup, review ReviewSummary) Reviewer {
	defer group.incPos()

	for len(group.Reviewers) > 1 &&
		(group.getCurrentReviewer().Alias == review.AuthorAlias ||
			group.getCurrentReviewer().VisualStudioID == review.AuthorVstsID) {

		group.incPos()
	}

	return group.getCurrentReviewer()
}
