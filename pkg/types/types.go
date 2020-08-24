package types

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// GraphUser stores information retrieved from the graph API
type GraphUser struct {
	OdataContext      string      `json:"@odata.context"`
	BusinessPhones    []string    `json:"businessPhones"`
	DisplayName       string      `json:"displayName"`
	GivenName         string      `json:"givenName"`
	JobTitle          string      `json:"jobTitle"`
	Mail              string      `json:"mail"`
	MobilePhone       interface{} `json:"mobilePhone"`
	OfficeLocation    string      `json:"officeLocation"`
	PreferredLanguage interface{} `json:"preferredLanguage"`
	Surname           string      `json:"surname"`
	UserPrincipalName string      `json:"userPrincipalName"`
	ID                string      `json:"id"`
}

// Repository holds the information for a repository
type Repository struct {
	ID             bson.ObjectId  `json:"id,omitempty" bson:"_id,omitempty"`
	Created        *time.Time     `json:"created,omitempty" bson:"_created,omitempty"`
	Updated        *time.Time     `json:"updated,omitempty" bson:"_updated,omitempty"`
	Name           string         `json:"name" bson:"name,omitempty"`
	ProjectName    string         `json:"projectName" bson:"projectName,omitempty"`
	ReviewerGroups ReviewerGroups `json:"reviewerGroups" bson:"reviewerGroups,omitempty"`
	Enabled        bool           `json:"enabled" bson:"enabled,omitempty"`
	Owners         []string       `json:"owners" bson:"owners,omitempty"`
	AdoRepoID      string         `json:"AdoRepoID" bson: "AdoRepoID,omitempty"`
}

// BaseGroup holds the base groups to be added or removed from a repo
type BaseGroup struct {
	ID             bson.ObjectId  `json:"_id,omitempty" bson:"_id,omitempty"`
	Created        *time.Time     `json:"_created,omitempty" bson:"_created,omitempty"`
	Updated        *time.Time     `json:"_updated,omitempty" bson:"_updated,omitempty"`
	Name           string         `json:"name" bson:"name,omitempty"`
	ReviewerGroups ReviewerGroups `json:"reviewerGroups" bson:"reviewerGroups,omitempty"`
	Owners         []string       `json:"owners" bson:"owners,omitempty"`
}

// ReviewerGroups is a list of type ReviewerGroup
type ReviewerGroups map[string]*ReviewerGroup

// ReviewerPositions holds the current position information for the reviewers
type ReviewerPositions map[string]int

// ReviewerGroup holds the reviwers and metadata for a review group.
type ReviewerGroup struct {
	Group      string      `json:"group" bson:"group,omitempty"`
	Required   bool        `json:"required" bson:"required,omitempty"`
	Reviewers  []*Reviewer `json:"reviewers" bson:"reviewers,omitempty"`
	CurrentPos int         `json:"currentPos" bson:"currentPos,omitempty"`
}

// Reviewer is a vsts revier object
type Reviewer struct {
	UniqueName string `json:"uniqueName" bson:"uniqueName,omitempty"`
	Alias      string `json:"alias" bson:"alias,omitempty"`
	ID         string `json:"id" bson:"id,omitempty"`
}

func (g *ReviewerGroup) getCurrentReviewer() *Reviewer {
	return g.Reviewers[g.CurrentPos]
}

func (g *ReviewerGroup) incPos() {
	g.CurrentPos = (g.CurrentPos + 1) % len(g.Reviewers)
}

// GetReviewers gets the required and optional reviewers for a review
// review: the review summary
// return: returns a slice of require reviewers and a slice of optional reviewers
func (rg *ReviewerGroups) GetReviewers(pullRequestCreatorID string) ([]*Reviewer, []*Reviewer, error) {
	requiredReviewers := make([]*Reviewer, 0, len(*rg)/2)
	optionalReviewers := make([]*Reviewer, 0, len(*rg)/2)

	for _, reviewerGroup := range *rg {
		if reviewerGroup.Required == true {
			requiredReviewers = append(requiredReviewers, getNextReviewer(reviewerGroup, pullRequestCreatorID))
		} else {
			optionalReviewers = append(optionalReviewers, getNextReviewer(reviewerGroup, pullRequestCreatorID))
		}
	}

	return requiredReviewers, optionalReviewers, nil
}

func getNextReviewer(group *ReviewerGroup, pullRequestCreatorID string) *Reviewer {
	defer group.incPos()

	for len(group.Reviewers) > 1 && group.getCurrentReviewer().ID == pullRequestCreatorID {

		group.incPos()
	}

	return group.getCurrentReviewer()
}
