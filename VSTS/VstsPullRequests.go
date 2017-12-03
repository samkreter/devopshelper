package vsts

import (
	"github.com/samkreter/VSTSAutoReviewer/review"
)

type VstsPullRequests struct {
	PullRequests []VstsPullRequest `json:"value"`
}

type VstsPullRequest struct {
	Id         string          `json:"pullRequestId"`
	Author     review.Reviewer `json:"createdBy"`
	Repository VstsRepository  `json:"repository"`
}

type VstsRepository struct {
	Id string `json:"id"`
}

func NewReviewSummary(pullRequest VstsPullRequest) review.ReviewSummary {
	return review.ReviewSummary{
		Id:           pullRequest.Id,
		AuthorEmail:  pullRequest.Author.Email,
		AuthorVstsId: pullRequest.Author.VisualStudioId,
		RepositoryId: pullRequest.Repository.Id,
		ReviewType:   "VstsPullRequest"}
}
