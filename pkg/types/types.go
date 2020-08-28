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
	Enabled        bool           `json:"enabled" bson:"enabled,omitempty"`
	AdoRepoID      string         `json:"AdoRepoID" bson: "AdoRepoID,omitempty"`
	LastReconciled time.Time
}

type Reviewer struct {
	Alias          string `json:"alias" bson:"alias,omitempty"`
	AdoID          string `json:"adoId" bson:"id,omitempty"`
	Id 			   bson.ObjectId `json:"id" bson:"_id,omitempty"`
	LastReviewTime time.Time
}
