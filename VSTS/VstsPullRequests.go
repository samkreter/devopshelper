package vsts

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
