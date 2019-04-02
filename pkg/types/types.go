package types

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/globalsign/mgo/bson"
)

// Repository holds the information for a repository
type Repository struct {
	ID             bson.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`
	Created        *time.Time    `json:"_created,omitempty" bson:"_created,omitempty"`
	Updated        *time.Time    `json:"_updated,omitempty" bson:"_updated,omitempty"`
	Name           string
	ProjectName    string
	ReviewerGroups ReviewerGroups
	CurrentPos     map[string]int
	Enabled        bool
}

// BaseGroup holds the base groups to be added or removed from a repo
type BaseGroup struct {
	ID             bson.ObjectId `json:"_id,omitempty" bson:"_id,omitempty"`
	Created        *time.Time    `json:"_created,omitempty" bson:"_created,omitempty"`
	Updated        *time.Time    `json:"_updated,omitempty" bson:"_updated,omitempty"`
	Name           string
	ReviewerGroups ReviewerGroups
}

// ReviewerGroups is a list of type ReviewerGroup
type ReviewerGroups map[string]*ReviewerGroup

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

// SavePositions saves the current position
func (rg *ReviewerGroups) SavePositions(statusFile string) error {
	reviewerPositions := make(ReviewerPositions)
	for _, reviewerGroup := range *rg {
		reviewerPositions[reviewerGroup.Group] = reviewerGroup.CurrentPos
	}

	data, err := json.Marshal(reviewerPositions)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(statusFile, data, 0644); err != nil {
		return err
	}

	log.Println("INFO: Saving position file.")
	return nil
}

func (g *ReviewerGroup) getCurrentReviewer() Reviewer {
	return g.Reviewers[g.CurrentPos]
}

func (g *ReviewerGroup) incPos() {
	g.CurrentPos = (g.CurrentPos + 1) % len(g.Reviewers)
}

// GetReviewers gets the required and optional reviewers for a review
// review: the review summary
// return: returns a slice of require reviewers and a slice of optional reviewers
func (rg *ReviewerGroups) GetReviewers(pullRequestCreatorID, statusFile string) ([]Reviewer, []Reviewer, error) {
	requiredReviewers := make([]Reviewer, 0, len(*rg)/2)
	optionalReviewers := make([]Reviewer, 0, len(*rg)/2)

	for _, reviewerGroup := range *rg {
		if reviewerGroup.Required == true {
			requiredReviewers = append(requiredReviewers, getNextReviewer(reviewerGroup, pullRequestCreatorID))
		} else {
			optionalReviewers = append(optionalReviewers, getNextReviewer(reviewerGroup, pullRequestCreatorID))
		}
	}

	if err := rg.SavePositions(statusFile); err != nil {
		return nil, nil, err
	}

	return requiredReviewers, optionalReviewers, nil
}

func getNextReviewer(group *ReviewerGroup, pullRequestCreatorID string) Reviewer {
	defer group.incPos()

	for len(group.Reviewers) > 1 && group.getCurrentReviewer().ID == pullRequestCreatorID {

		group.incPos()
	}

	return group.getCurrentReviewer()
}
