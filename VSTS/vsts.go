package vsts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type config struct {
	VstsToken         string `json:"vstsToken"`
	VstsProject       string `json:"vstsProject"`
	VstsUsername      string `json:"vstsUsername"`
	VstsRepositoryID  string `json:"repositoryId"`
	VstsArmReviewerID string `json:"vstsArmReviewerId"`
	VstsBotMaker      string `json:"vstsBotMaker"`
}

var (
	Conf                    *config
	PullRequestsURITemplate = "DefaultCollection/{project}/_apis/git/pullRequests?api-version={apiVersion}&reviewerId={reviewerId}"
	CommentsURITemplate     = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/threads?api-version={apiVersion}"
	ReviewerURITemplate     = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/reviewers/{reviewerId}?api-version={apiVersion}"
	VstsBaseURI             = "https://msazure.visualstudio.com/"
	APIVersion              = "3.0"
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

func GetCommentsUri(pullRequestID string, repositoryID string) string {
	r := strings.NewReplacer("{repositoryId}", repositoryID,
		"{pullRequestId", pullRequestID,
		"{apiVersion}", APIVersion)
	return fmt.Sprintf("%s%s", VstsBaseURI, r.Replace(ReviewerURITemplate))
}

func GetReviewerUri(repositoryID string, pullRequestID string, reviewerID string) string {
	r := strings.NewReplacer("{repositoryId}", repositoryID,
		"{pullRequestId", pullRequestID,
		"{reviewerId}", reviewerID,
		"{apiVersion}", APIVersion)
	return fmt.Sprintf("%s%s", VstsBaseURI, r.Replace(ReviewerURITemplate))
}

func GetPullRequestsUri() string {
	r := strings.NewReplacer("{project}", Conf.VstsProject,
		"{reviewerId}", Conf.VstsArmReviewerID,
		"{apiVersion}", APIVersion)
	return fmt.Sprintf("%s%s", VstsBaseURI, r.Replace(PullRequestsURITemplate))
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

func ContainsReviewBalancerComment(reviewSummary ReviewSummary) bool {
	url := GetCommentsUri(reviewSummary.RepositoryID, reviewSummary.ID)

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

func GetInprogressReviews() []ReviewSummary {
	url := GetPullRequestsUri()

	pullRequests := new(VstsPullRequests)
	err := GetJsonResponse(url, pullRequests)
	if err != nil {
		log.Fatal(err)
	}

	reviewSummaries := make([]ReviewSummary, len(pullRequests.PullRequests))
	for index, pullRequest := range pullRequests.PullRequests {
		reviewSummaries[index] = NewReviewSummary(pullRequest)
	}
	return reviewSummaries
}

func AddRootComment(reviewSummary ReviewSummary, comment string) {
	thread := NewVstsCommentThread(comment)

	url := GetCommentsUri(reviewSummary.RepositoryID, reviewSummary.ID)
	err := PostJson(url, thread)
	if err != nil {
		log.Fatal(err)
	}
}

func AddReviewers(reviewSummary ReviewSummary, required []Reviewer, optional []Reviewer) {
	for _, reviewer := range append(required, optional...) {
		url := GetReviewerUri(reviewSummary.RepositoryID, reviewSummary.ID, reviewer.VisualStudioID)
		vote := NewDefaultVisualStudioReviewerVote()

		err := PostJson(url, vote)
		if err != nil {
			log.Fatal(err)
		}
	}
}
