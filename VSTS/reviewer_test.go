package vsts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	actualReviewers = Reviewers{
		Required: []Reviewer{
			{
				VisualStudioId: "asdfalksjdfji33u34ii",
				Email:          "sakreter@microsoft.com",
				Alias:          "sakreter",
			},
		},
		Optional: []Reviewer{
			{
				VisualStudioId: "asddas33333ksjdfji33u34ii",
				Email:          "tesdad@microsoft.com",
				Alias:          "tesdad",
			},
		},
	}
)

func TestGetReviewersAlias(t *testing.T) {
	combinedReviewers := append(actualReviewers.Required, actualReviewers.Optional...)
	alias := GetReviewersAlias(combinedReviewers)

	assert.Equal(t, []string{"sakreter", "tesdad"}, alias)
}

func TestGetNextReviewers(t *testing.T){
	review := ReviewSummary{
		Id: "testiasdfasdf",
		AuthorVstsID: "asdfalksjdfji33u34ii",
		AuthorEmail:  "sakreter@microsoft.com",
		AuthorAlias:  "sakreter",
		RepositoryId: "112341234556623",
		ReviewType:   "basic",
	}
	
	//Base Test
	req, op := GetNextReviewers(review)
	assert.Equal(t, "dasdfe" , req[0].Alias)
	assert.Equal(t, "tesdad", op[0].Alias)

	//Should go to next reviewer
	req, op = GetNextReviewers(review)
	assert.Equal(t, "edfgaa" , req[0].Alias)
	assert.Equal(t, "yekk34t", op[0].Alias)

	//Should go to begining for required
	req, op = GetNextReviewers(review)
	assert.Equal(t, "dasdfe" , req[0].Alias)
	assert.Equal(t, "dfedad", op[0].Alias)

}

func TestLoadReviewers(t *testing.T) {
	expectedReviewers := Reviewers{
		Required: []Reviewer{
			{
				VisualStudioId: "asdfalksjdfji33u34ii",
				Email:          "sakreter@microsoft.com",
				Alias:          "sakreter",
			},
		},
		Optional: []Reviewer{
			{
				VisualStudioId: "asddas33333ksjdfji33u34ii",
				Email:          "tesdad@microsoft.com",
				Alias:          "tesdad",
			},
		},
	}
	actualReviewers := loadReviewers()

	assert.Equal(t, 3, len(actualReviewers.Required), "Must have the correct length of required reviewers")
	assert.Equal(t, 3, len(actualReviewers.Required), "Must have the correct length of optional reviewersd")
	assert.Equal(t, expectedReviewers.Required[0], actualReviewers.Required[0])
	assert.Equal(t, expectedReviewers.Optional[0], actualReviewers.Optional[0])
}
