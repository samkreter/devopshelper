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
	Token          string `json:"vstsToken"`
	Project        string `json:"vstsProject"`
	Username       string `json:"vstsUsername"`
	RepositoryName string `json:"repositoryName"`
	ArmReviewerID  string `json:"vstsArmReviewerId"`
	APIVersion     string `json:"vstsApiVersion"`
	BotMaker       string `json:"vstsBotMaker"`
}

var (
	Config                  = vstsConfig{}
	PullRequestsURITemplate = "DefaultCollection/{project}/_apis/git/repositories/{repositoryName}/pullRequests?api-version={apiVersion}"
	CommentsURITemplate     = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/threads?api-version={apiVersion}"
	ReviewerURITemplate     = "DefaultCollection/{project}/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/reviewers/{reviewerId}?api-version={apiVersion}"
	VstsBaseURI             = "https://msazure.visualstudio.com/"
	APIVersion              = "3.0"
)

func init() {
	configor.Load(&Config, "configs/config.dev.json")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func GetCommentsUri(repositoryID string, pullRequestID string) string {
	r := strings.NewReplacer(
		"{repositoryId}", repositoryID,
		"{pullRequestId}", pullRequestID,
		"{apiVersion}", APIVersion)
	return fmt.Sprintf("%s%s", VstsBaseURI, r.Replace(CommentsURITemplate))
}

func GetReviewerUri(repositoryID string, pullRequestID string, reviewerID string) string {
	r := strings.NewReplacer(
		"{project}", Config.Project,
		"{repositoryId}", repositoryID,
		"{pullRequestId}", pullRequestID,
		"{reviewerId}", reviewerID,
		"{apiVersion}", APIVersion)
	return fmt.Sprintf("%s%s", VstsBaseURI, r.Replace(ReviewerURITemplate))
}

func GetPullRequestsUri() string {
	r := strings.NewReplacer(
		"{project}", Config.Project,
		"{repositoryName}", Config.RepositoryName,
		"{apiVersion}", APIVersion)

	return fmt.Sprintf("%s%s", VstsBaseURI, r.Replace(PullRequestsURITemplate))
}

func SendJson(method string, url string, jsonData interface{}) error {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(jsonData)

	req, err := http.NewRequest(method, url, b)
	req.SetBasicAuth(Config.Username, Config.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("SendJson: repsonse with non 200 code of %d", resp.StatusCode)
	}

	return nil
}

func GetJsonResponse(url string, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(Config.Username, Config.Token)

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
				if strings.Contains(comment.Content, Config.BotMaker) {
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
	err := SendJson("POST", url, thread)
	if err != nil {
		log.Fatal(err)
	}
}

func AddReviewers(reviewSummary ReviewSummary, required []Reviewer, optional []Reviewer) {
	for _, reviewer := range append(required, optional...) {
		url := GetReviewerUri(reviewSummary.RepositoryID, reviewSummary.ID, reviewer.VisualStudioID)

		vote := NewDefaultVisualStudioReviewerVote()

		fmt.Println("url", url)

		err := SendJson("PUT", url, vote)
		if err != nil {
			log.Fatal(err)
		}
	}
}
