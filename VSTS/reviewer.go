package vsts

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type ReviewerGroups []ReviewerGroup

type ReviewerGroup struct {
	Group      string     `json:"group"`
	Required   bool       `json:"required"`
	Reviewers  []Reviewer `json:"reviewers"`
	CurrentPos int
}

type Reviewer struct {
	VisualStudioId string `json:"id"`
	Email          string `json:"uniqueName"`
	Alias          string `json:"alias"`
}

func (g ReviewerGroup) getCurrentReviewer() Reviewer {
	return g.Reviewers[g.CurrentPos]
}

func (g *ReviewerGroup) incPos() {
	g.CurrentPos = (g.CurrentPos + 1) % len(g.Reviewers)
}

type ReviewSummary struct {
	Id           string
	AuthorAlias  string
	AuthorEmail  string
	AuthorVstsID string
	RepositoryId string
	ReviewType   string
}

var (
	reviewerGroups ReviewerGroups
)

func init() {
	reviewerGroups = loadReviewerGroups()
}

func GetReviewersAlias(reviewers []Reviewer) []string {
	aliases := make([]string, len(reviewers))

	for index, reviewer := range reviewers {
		aliases[index] = reviewer.Alias
	}
	return aliases
}

func loadReviewerGroups() ReviewerGroups {
	rawData, err := ioutil.ReadFile("./reviewers.json")
	if err != nil {
		log.Fatal(err)
	}

	reviewerGroups := ReviewerGroups{}
	json.Unmarshal(rawData, &reviewerGroups)

	return reviewerGroups
}

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

	return requiredReviewers, optionalReviewers
}

func getNextReviewer(group *ReviewerGroup, review ReviewSummary) Reviewer {
	defer group.incPos()

	for len(group.Reviewers) > 1 &&
		(group.getCurrentReviewer().Alias == review.AuthorAlias ||
			group.getCurrentReviewer().VisualStudioId == review.AuthorVstsID) {

		group.incPos()
	}

	return group.getCurrentReviewer()
}
