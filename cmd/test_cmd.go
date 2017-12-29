package cmd

import (
	"github.com/samkreter/VSTSAutoReviewer/vsts"
)

func RunTest() error {
	review := vsts.ReviewSummary{
		AuthorEmail:  "sakreter@microsoft.com",
		AuthorVstsID: "3a883e38-d0ad-62e5-aff2-e30dca0db57e",
		ID:           "516639",
		RepositoryID: "bca7b01a-2f98-4902-ad02-e0fd81a82ea1",
	}
	vsts.AddRootComment(review, "Testing Review Tester")
	return nil
}
