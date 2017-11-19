package review

import (
	"log"
	"io/ioutil"
	"encoding/json"
)

type Reviewers struct{
	optional []Reviewer
	required []Reviewer
}

type Reviewer struct{
	VisualStudioId string `json:"id"`
	Email string `json:"uniqueName"`
	Alias string `json:"alias"`
}

func GetReviewersAlias(reviewers []Reviewer) []string{
	aliases := make([]string,len(reviewers))
	
	for index, reviewer := range reviewers{
		aliases[index] = reviewer.Alias
	}
	return aliases
}

func LoadReviewers() ([]Reviewer,[]Reviewer){
	rawData, err := ioutil.ReadFile("./reviewers.json")
    if err != nil {
		log.Fatal(err)
    }

    var reviewers Reviewers
    json.Unmarshal(rawData, &reviewers)
	
	return reviewers.required,reviewers.optional
}