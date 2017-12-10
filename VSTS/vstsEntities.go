package vsts

//Pull Requests
type VstsPullRequests struct {
	PullRequests []VstsPullRequest `json:"value"`
}

type VstsPullRequest struct {
	ID         string         `json:"pullRequestId"`
	Author     Reviewer       `json:"createdBy"`
	Repository VstsRepository `json:"repository"`
}

type VstsRepository struct {
	ID string `json:"id"`
}

func NewReviewSummary(pullRequest VstsPullRequest) ReviewSummary {
	return ReviewSummary{
		ID:           pullRequest.ID,
		AuthorEmail:  pullRequest.Author.Email,
		AuthorVstsID: pullRequest.Author.VisualStudioID,
		RepositoryID: pullRequest.Repository.ID,
		ReviewType:   "VstsPullRequest"}
}

//Comment Thread
type VstsCommentThreads struct {
	CommentThreads []VstsCommentThread `json:"value"`
}

type VstsCommentThread struct {
	Comments []VstsComment `json:"comments"`
}

type VstsComment struct {
	Content string `json:"content"`
}

func NewVstsCommentThread(comment string) VstsCommentThread {
	return VstsCommentThread{
		Comments: []VstsComment{VstsComment{Content: comment}}}
}

// Reviewer Vote
const (
	Rejected                = -10
	Waiting                 = -5
	NoResponse              = 0
	ApprovedWithSuggestions = 5
	Approved                = 10
)

type VstsReviewerVote struct {
	Vote int `json:"vote"`
}

func NewDefaultVisualStudioReviewerVote() VstsReviewerVote {
	return VstsReviewerVote{Vote: NoResponse}
}
