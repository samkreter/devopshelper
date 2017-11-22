package main

import (
	"fmt"
	"strings"
	"github.com/samkreter/VSTSAutoReviewer/vsts"
	"github.com/samkreter/VSTSAutoReviewer/review"
)

func main(){
	fmt.Println("hello")
}

func CheckReviews(){
	reviews := vsts.GetInprogressReviews()

	//Todo: add codeflow reviews

	for _, review := range reviews {
		BalanceReview(review)
	}

}

func BalanceReview(reviewSummary vsts.ReviewSummary){
	if (!vsts.ContainsReviewBalancerComment(reviewSummary)){
		
		//need reiew iteration algo to change reviewers
		requiredReviewers, optionalReviewers := review.LoadReviewers()

		vsts.AddReviewers(reviewSummary, requiredReviewers, optionalReviewers)

		comment := fmt.Sprintf(
		"Hello %s,\r\n\r\n" +
		"You are randomly selected as the **required** code reviewers of this change. \r\n\r\n" +
		"Your responsibility is to review **each** iteration of this CR until signoff. You should provide no more than 48 hour SLA for each iteration.\r\n\r\n" +
		"Thank you.\r\n\r\n" +
		"CR Balancer\r\n" +
		"%s",
		strings.Join(review.GetReviewersAlias(requiredReviewers),","),
		vsts.Conf.VstsBotMaker);


		vsts.AddRootComment(reviewSummary, comment)
	}
}
