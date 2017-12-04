package vsts

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var (
	reviewers   *Reviewers
)

type Reviewers struct {
	Optional    []Reviewer `json:"optional"`
	Required    []Reviewer `json:"required"`
	requiredPos int
	optionalPos int
}

func (r Reviewers) getCurrentRequred() Reviewer {
	return r.Required[r.requiredPos]
}

func (r Reviewers) getCurrentOptional() Reviewer {
	return r.Optional[r.optionalPos]
}

func (r *Reviewers) incRequred() {
	r.requiredPos = (r.requiredPos + 1) % len(r.Required)
}

func (r *Reviewers) incOptional() {
	r.optionalPos = (r.optionalPos + 1) % len(r.Optional)
}

type ReviewSummary struct {
	Id           string
	AuthorAlias  string
	AuthorEmail  string
	AuthorVstsID string
	RepositoryId string
	ReviewType   string
}

type Reviewer struct {
	VisualStudioId string `json:"id"`
	Email          string `json:"uniqueName"`
	Alias          string `json:"alias"`
}

func init() {
	reviewers = loadReviewers()
}

func GetReviewersAlias(reviewers []Reviewer) []string {
	aliases := make([]string, len(reviewers))

	for index, reviewer := range reviewers {
		aliases[index] = reviewer.Alias
	}
	return aliases
}

func loadReviewers() *Reviewers {
	rawData, err := ioutil.ReadFile("./reviewers.json")
	if err != nil {
		log.Fatal(err)
	}

	reviewers := &Reviewers{}
	json.Unmarshal(rawData, &reviewers)

	return reviewers
}

func GetNextReviewers(review ReviewSummary) ([]Reviewer, []Reviewer) {
	defer reviewers.incOptional()
	defer reviewers.incRequred()

	for len(reviewers.Required) > 1 && 
		(reviewers.getCurrentRequred().Alias == review.AuthorAlias ||
		reviewers.getCurrentRequred().VisualStudioId == review.AuthorVstsID) {
		
		reviewers.incRequred()
	}

	for len(reviewers.Optional) > 1 && 
		(reviewers.getCurrentOptional().Alias == review.AuthorAlias ||
		reviewers.getCurrentOptional().VisualStudioId == review.AuthorVstsID) {
		
		reviewers.incOptional()
	}

	return []Reviewer{reviewers.getCurrentRequred()}, []Reviewer{reviewers.getCurrentOptional()}
}
