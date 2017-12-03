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
		Id:           pullRequest.ID,
		AuthorEmail:  pullRequest.Author.Email,
		AuthorVstsID: pullRequest.Author.VisualStudioId,
		RepositoryId: pullRequest.Repository.ID,
		ReviewType:   "VstsPullRequest"}
}
