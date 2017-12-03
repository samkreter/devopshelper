package vsts

import (
	"fmt"
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

func TestLoadReviewers(t *testing.T) {
	actualReviewers := Reviewers{
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
	requried, optional := LoadReviewers()

	fmt.Println(requried)

	assert.Equal(t, actualReviewers.Required[0], requried[0])
	assert.Equal(t, actualReviewers.Optional[0], optional[0])
}
