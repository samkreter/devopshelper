package review

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type Reviewers struct {
	Optional []Reviewer `json:"optional"`
	Required []Reviewer `json:"required"`
}

type ReviewSummary struct {
	Id           string
	AuthorAlias  string
	AuthorEmail  string
	AuthorVstsId string
	RepositoryId string
	ReviewType   string
}

type Reviewer struct {
	VisualStudioId string `json:"id"`
	Email          string `json:"uniqueName"`
	Alias          string `json:"alias"`
}

func GetReviewersAlias(reviewers []Reviewer) []string {
	aliases := make([]string, len(reviewers))

	for index, reviewer := range reviewers {
		aliases[index] = reviewer.Alias
	}
	return aliases
}

func LoadReviewers() ([]Reviewer, []Reviewer) {
	rawData, err := ioutil.ReadFile("./reviewers.json")
	if err != nil {
		log.Fatal(err)
	}

	var reviewers Reviewers
	json.Unmarshal(rawData, &reviewers)

	return reviewers.Required, reviewers.Optional
}

func GetNextReviewers(reveiw ReviewSummary) ([]Reviewer, []Reviewer) {
	fmt.Println(Conf)
	log.Fatal("Not Implemented")
	return nil, nil
}
