package vsts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/configor"
)

type vstsConfig struct {
	VstsToken          string `json:"vstsToken"`
	VstsProject        string `json:"vstsProject"`
	VstsUsername       string `json:"vstsUsername"`
	VstsRepositoryName string `json:"repositoryName"`
	VstsArmReviewerID  string `json:"vstsArmReviewerId"`
	VstsAPIVersion     string `json:"vstsApiVersion"`
	VstsBotMaker       string `json:"vstsBotMaker"`
}

var (
	Config                  = vstsConfig{}
	PullRequestsURITemplate = "DefaultCollection/{project}/_apis/git/repositories/{repositoryName}/pullRequests?api-version={apiVersion}"
	CommentsURITemplate     = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/threads?api-version={apiVersion}"
	ReviewerURITemplate     = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/reviewers/{reviewerId}?api-version={apiVersion}"
	VstsBaseURI             = "https://msazure.visualstudio.com/"
	APIVersion              = "3.0"
)

func init() {
	configor.Load(&Config, "configs/config.dev.json")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func GetCommentsUri(repositoryID string, pullRequestID string) string {
	r := strings.NewReplacer("{repositoryId}", repositoryID,
		"{pullRequestId}", pullRequestID,
		"{apiVersion}", APIVersion)
	return fmt.Sprintf("%s%s", VstsBaseURI, r.Replace(CommentsURITemplate))
}

func GetReviewerUri(repositoryID string, pullRequestID string, reviewerID string) string {
	r := strings.NewReplacer("{repositoryId}", repositoryID,
		"{pullRequestId", pullRequestID,
		"{reviewerId}", reviewerID,
		"{apiVersion}", APIVersion)
	return fmt.Sprintf("%s%s", VstsBaseURI, r.Replace(ReviewerURITemplate))
}

func GetPullRequestsUri() string {
	r := strings.NewReplacer("{project}", Config.VstsProject,
		"{repositoryName}", Config.VstsRepositoryName,
		"{apiVersion}", APIVersion)

	return fmt.Sprintf("%s%s", VstsBaseURI, r.Replace(PullRequestsURITemplate))
}

func PostJson(url string, jsonData interface{}) error {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(jsonData)

	req, err := http.NewRequest("POST", url, b)
	req.SetBasicAuth(Config.VstsUsername, Config.VstsToken)
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
	req.SetBasicAuth(Config.VstsUsername, Config.VstsToken)

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
				if strings.Contains(comment.Content, Config.VstsBotMaker) {
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
