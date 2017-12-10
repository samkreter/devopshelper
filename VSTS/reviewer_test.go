package vsts

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	out                    io.Writer = os.Stdout
	expectedReviewerGroups           = ReviewerGroups{
		{
			Group:    "Tier1",
			Required: true,
			Reviewers: []Reviewer{
				{
					VisualStudioID: "asdfalksjdfji33u34ii",
					Email:          "sakreter@microsoft.com",
					Alias:          "sakreter",
				},
				{
					VisualStudioID: "asdfalkd345d3u34ii",
					Email:          "dasdfe@microsoft.com",
					Alias:          "dasdfe",
				},
				{
					VisualStudioID: "aasda3333eefgii",
					Email:          "edfgaa@microsoft.com",
					Alias:          "edfgaa",
				},
			},
			CurrentPos: 0,
		},
	}
)

func l(output string) {
	fmt.Fprintf(out, output)
}
func TestGetReviewersAlias(t *testing.T) {
	alias := GetReviewersAlias(expectedReviewerGroups[0].Reviewers)

	assert.Equal(t, []string{"sakreter", "dasdfe", "edfgaa"}, alias)
}

func TestGetReviewers(t *testing.T) {
	review := ReviewSummary{
		ID:           "testiasdfasdf",
		AuthorVstsID: "asdfalksjdfji33u34ii",
		AuthorEmail:  "sakreter@microsoft.com",
		AuthorAlias:  "sakreter",
		RepositoryID: "112341234556623",
		ReviewType:   "basic",
	}

	//Base Test
	req, op := GetReviewers(review)
	assert.Equal(t, 2, len(req))
	assert.Equal(t, 1, len(op))
	assert.Equal(t, "dasdfe", req[0].Alias)
	assert.Equal(t, "psaraiya", op[0].Alias)

	//Should go to next reviewer
	req, op = GetReviewers(review)
	assert.Equal(t, 2, len(req))
	assert.Equal(t, 1, len(op))
	assert.Equal(t, "edfgaa", req[0].Alias)
	assert.Equal(t, "psaraiya", op[0].Alias)

	///Should go to begining for required
	req, op = GetReviewers(review)
	assert.Equal(t, 2, len(req))
	assert.Equal(t, 1, len(op))
	assert.Equal(t, "dasdfe", req[0].Alias)
	assert.Equal(t, "psaraiya", op[0].Alias)
}

func TestLoadReviewerGroups(t *testing.T) {
	actualReviewerGroups := loadReviewerGroups()

	fmt.Fprint(os.Stdout, "testing")
	fmt.Fprint(out, actualReviewerGroups[0])

	assert.Equal(t, 6, len(actualReviewerGroups), "Must have the correct length of reviewerGroup")

	assert.Equal(t, expectedReviewerGroups[0].Reviewers[0], actualReviewerGroups[0].Reviewers[0])
	assert.Equal(t, expectedReviewerGroups[0].Required, "true")
}
