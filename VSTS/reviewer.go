package vsts

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var (
	requiredPos = 0
	optionalPos = 0
	reviewers   *Reviewers
)

type Reviewers struct {
	Optional    []Reviewer `json:"optional"`
	Required    []Reviewer `json:"required"`
	requiredPos int
	optionalPos int
}

func getCurrentRequred() Reviewer {
	return reviewers.Required[requiredPos]
}

func getCurrentOptional() Reviewer {
	return reviewers.Optional[optionalPos]
}

func incRequred() {
	reviewers.requiredPos++
}

func incOptional() {
	reviewers.optionalPos++
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

	reviewers.requiredPos = 0
	reviewers.optionalPos = 0

	return reviewers
}

func GetNextReviewers(review ReviewSummary) (Reviewer, Reviewer) {
	requiredPos++
	optionalPos++

	for reviewers.Required[requiredPos].Alias == review.AuthorAlias ||
		reviewers.Required[requiredPos].Alias == review.AuthorVstsID {

		requiredPos++
	}

	for reviewers.Optional[requiredPos].Alias == review.AuthorAlias ||
		reviewers.Optional[requiredPos].Alias == review.AuthorVstsID {

		optionalPos++
	}

	return reviewers.Required[requiredPos], reviewers.Optional[optionalPos]
}
