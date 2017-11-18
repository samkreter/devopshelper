package VSTS  


type ReviewSummary struct{
	Id 				string
	AuthorAlias 	string 
    AuthorEmail 	string
    AuthorVstsId 	string
	RepositoryId 	string
	ReviewType		string
}

func NewReviewSummary(pullRequest VstsPullRequest) ReviewSummary {
	return ReviewSummary{
		Id: pullRequest.Id,
		AuthorEmail: pullRequest.Author.Email,
		AuthorVstsId: pullRequest.Author.VisualStudioId,
		RepositoryId: pullRequest.Repository.Id,
		ReviewType: "VstsPullRequest"}
}