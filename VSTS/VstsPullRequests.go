package VSTS

type VstsPullRequests struct{
	PullRequests []VstsPullRequest `json:"value"`
}

type VstsPullRequest struct
{
	Id 			string 					`json:"pullRequestId"`
	Author 		Reviewer  				`json:"createdBy"`
	Repository 	VstsRepository  `json:"repository"`
}

type VstsRepository struct{
	Id string `json:"id"`
}

type Reviewer struct{
	VisualStudioId string `json:"id"`
	Email string `json:"uniqueName"`
	Alias string `json:"alias"`
}