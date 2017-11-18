package VSTS  


type ReviewSummary struct{
	Id 				string
	AuthorAlias 	string 
    AuthorEmail 	string
    AuthorVstsId 	string
	RepositoryId 	string
	ReviewType		string
}

func (reviewSummary *ReviewSummary) GetReviewSummary(pullRequest VstsPullRequest){
	reviewSummary.Id = pullRequest.Id
	reviewSummary.AuthorEmail = pullRequest.Author.Email
	reviewSummary.AuthorVstsId = pullRequest.Author.VisualStudioId
	reviewSummary.RepositoryId = pullRequest.Repository.Id
	reviewSummary.ReviewType = "VstsPullRequest"
}