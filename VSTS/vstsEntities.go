package vsts

import (
	"encoding/json"
)

//PullRequests is a list of pull requests.
type PullRequests struct {
	PullRequests []PullRequest `json:"value"`
}

//PullRequest is a pull request retrieved from the vsts API.
type PullRequest struct {
	ID         json.Number `json:"pullRequestId"`
	Author     Reviewer    `json:"createdBy"`
	Repository Repository  `json:"repository"`
}

// Repository is an identifier for a vsts repository.
type Repository struct {
	ID string `json:"id"`
}

// NewReviewSummary creates an interal review summary from a vsts pull request
func NewReviewSummary(pullRequest PullRequest) ReviewSummary {
	return ReviewSummary{
		ID:           string(pullRequest.ID),
		AuthorAlias:  pullRequest.Author.Alias,
		AuthorEmail:  pullRequest.Author.Email,
		AuthorVstsID: pullRequest.Author.VisualStudioID,
		RepositoryID: pullRequest.Repository.ID,
		ReviewType:   "VstsPullRequest"}
}

//CommentThreads holds a list of commentThreads
type CommentThreads struct {
	CommentThreads []CommentThread `json:"value"`
}

//CommentThread holds all comments for a vsts pull request
type CommentThread struct {
	Comments []Comment `json:"comments"`
}

//Comment is a single comment on a vsts pull request
type Comment struct {
	Content string `json:"content"`
}

//NewCommentThread creates a net comment thread from a single comment
func NewCommentThread(comment string) CommentThread {
	return CommentThread{
		Comments: []Comment{Comment{Content: comment}}}
}

// Reviewer Vote
const (
	Rejected                = -10
	Waiting                 = -5
	NoResponse              = 0
	ApprovedWithSuggestions = 5
	Approved                = 10
)

// ReviewerVote holds the vote item for the vsts API
type ReviewerVote struct {
	Vote int `json:"vote"`
}

// NewDefaultReviewerVote creates a new reviewer value with no response set.
func NewDefaultReviewerVote() ReviewerVote {
	return ReviewerVote{Vote: NoResponse}
}
