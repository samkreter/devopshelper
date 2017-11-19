package vsts

const (
	Rejected = -10
	Waiting = -5
	NoResponse = 0
	ApprovedWithSuggestions = 5
	Approved = 10
)

type VstsReviewerVote struct{
	Vote 	int 	`json:"vote"`
}

func NewDefaultVisualStudioReviewerVote() VstsReviewerVote{
	return VstsReviewerVote { Vote: NoResponse };
}
