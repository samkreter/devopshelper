package VSTS

type VisualStudioPullRequests struct{
	PullRequests []VisualStudioPullRequest `json:"value"`
}

type VisualStudioPullRequest struct
{
	Id 			string 					`json:"pullRequestId"`
	Author 		Reviewer  				`json:"createdBy"`
	Repository 	VisualStudioRepository  `json:"repository"`
}

type VisualStudioRepository struct{
	Id string `json:"id"`
}

type Reviewer struct{
	VisualStudioId string `json:"id"`
	Email string `json:"uniqueName"`
	Alias string `json:"alias"`
}