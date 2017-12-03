package vsts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/samkreter/VSTSAutoReviewer/review"
	"github.com/spf13/viper"
)

type config struct {
	VstsToken         string `json:"vstsToken"`
	VstsProject       string `json:"vstsProject"`
	VstsUsername      string `json:"vstsUsername"`
	VstsRepositoryId  string `json:"repositoryId"`
	VstsArmReviewerId string `json:"vstsArmReviewerId"`
	VstsBotMaker      string `json:"vstsBotMaker"`
}

var (
	Conf                    *config
	PullRequestsUriTemplate string = "DefaultCollection/{project}/_apis/git/pullRequests?api-version={apiVersion}&reviewerId={reviewerId}"
	CommentsUriTemplate     string = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/threads?api-version={apiVersion}"
	ReviewerUriTemplate     string = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/reviewers/{reviewerId}?api-version={apiVersion}"
	VstsBaseUri             string = "https://msazure.visualstudio.com/"
	ApiVersion              string = "3.0"
)

func getConf() *config {
	viper.AddConfigPath(".")
	viper.SetConfigName("config.dev")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("%v", err)
	}

	conf := &config{}
	err = viper.Unmarshal(conf)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}
	return conf
}

func init() {
	Conf = getConf()
}

func GetCommentsUri(pullRequestId string, repositoryId string) string {
	r := strings.NewReplacer("{repositoryId}", repositoryId,
		"{pullRequestId", pullRequestId,
		"{apiVersion}", ApiVersion)
	return fmt.Sprintf("%s%s", VstsBaseUri, r.Replace(ReviewerUriTemplate))
}

func GetReviewerUri(repositoryId string, pullRequestId string, reviewerId string) string {
	r := strings.NewReplacer("{repositoryId}", repositoryId,
		"{pullRequestId", pullRequestId,
		"{reviewerId}", reviewerId,
		"{apiVersion}", ApiVersion)
	return fmt.Sprintf("%s%s", VstsBaseUri, r.Replace(ReviewerUriTemplate))
}

func GetPullRequestsUri() string {
	r := strings.NewReplacer("{project}", Conf.VstsProject,
		"{reviewerId}", Conf.VstsArmReviewerId,
		"{apiVersion}", ApiVersion)
	return fmt.Sprintf("%s%s", VstsBaseUri, r.Replace(PullRequestsUriTemplate))
}

func PostJson(url string, jsonData interface{}) error {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(jsonData)

	req, err := http.NewRequest("POST", url, b)
	req.SetBasicAuth(Conf.VstsUsername, Conf.VstsToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("PostJson: repsonse with non 200 code of %d", resp.StatusCode)
	}

	return nil
}

func GetJsonResponse(url string, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(Conf.VstsUsername, Conf.VstsToken)

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(target)
}

func ContainsReviewBalancerComment(reviewSummary review.ReviewSummary) bool {
	url := GetCommentsUri(reviewSummary.RepositoryId, reviewSummary.Id)

	threads := new(VstsCommentThreads)
	err := GetJsonResponse(url, threads)
	if err != nil {
		log.Fatal(err)
	}

	if threads != nil {
		for _, thread := range threads.CommentThreads {
			for _, comment := range thread.Comments {
				if strings.Contains(comment.Content, Conf.VstsBotMaker) {
					return true
				}
			}
		}
	}
	return false
}

func GetInprogressReviews() []review.ReviewSummary {
	url := GetPullRequestsUri()

	pullRequests := new(VstsPullRequests)
	err := GetJsonResponse(url, pullRequests)
	if err != nil {
		log.Fatal(err)
	}

	reviewSummaries := make([]review.ReviewSummary, len(pullRequests.PullRequests))
	for index, pullRequest := range pullRequests.PullRequests {
		reviewSummaries[index] = NewReviewSummary(pullRequest)
	}
	return reviewSummaries
}

func AddRootComment(reviewSummary review.ReviewSummary, comment string) {
	thread := NewVstsCommentThread(comment)

	url := GetCommentsUri(reviewSummary.RepositoryId, reviewSummary.Id)
	err := PostJson(url, thread)
	if err != nil {
		log.Fatal(err)
	}
}

func AddReviewers(reviewSummary review.ReviewSummary, required []review.Reviewer, optional []review.Reviewer) {
	for _, reviewer := range append(required, optional...) {
		url := GetReviewerUri(reviewSummary.RepositoryId, reviewSummary.Id, reviewer.VisualStudioId)
		vote := NewDefaultVisualStudioReviewerVote()

		err := PostJson(url, vote)
		if err != nil {
			log.Fatal(err)
		}
	}
}
