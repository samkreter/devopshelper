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

type vstsConfig struct {
	Token          string `json:"token"`
	Project        string `json:"project"`
	Username       string `json:"username"`
	RepositoryName string `json:"repositoryName"`
	APIVersion     string `json:"apiVersion"`
	BotMaker       string `json:"botMaker"`
}

var (
	//Config holds the vsts configuration.
	Config                  = vstsConfig{}
	pullRequestsURITemplate = "DefaultCollection/{project}/_apis/git/repositories/{repositoryName}/pullRequests?api-version={apiVersion}&targetRefName=refs/heads/master"
	commentsURITemplate     = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/threads?api-version={apiVersion}"
	reviewerURITemplate     = "DefaultCollection/{project}/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/reviewers/{reviewerId}?api-version={apiVersion}"
	vstsBaseURI             = "https://msazure.visualstudio.com/"
)

func readConfig(filename string, envPrefix string, defaults map[string]interface{}) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}

	v.SetConfigName(filename)
	v.AddConfigPath("/configs")
	//v.AddConfigPath("./configs")
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}

func checkConfig(config vstsConfig) error {
	configError := "Must provide configuration for %s."

	if config.Project == "" {
		return fmt.Errorf(configError, "the VSTS project")
	}

	if config.RepositoryName == "" {
		return fmt.Errorf(configError, "the VSTS repository name")
	}

	if config.Token == "" {
		return fmt.Errorf(configError, "the VSTS personal token")
	}

	return nil
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	v, err := readConfig("vsts.config.dev", "vsts", map[string]interface{}{
		"APIVersion": "3.0",
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := v.Unmarshal(&Config); err != nil {
		log.Fatal(err)
	}

	if err := checkConfig(Config); err != nil {
		log.Fatal(err)
	}
}

// GetCommentsURI constructs the URI for the comments API interactions.
func GetCommentsURI(repositoryID string, pullRequestID string) string {
	r := strings.NewReplacer(
		"{repositoryId}", repositoryID,
		"{pullRequestId}", pullRequestID,
		"{apiVersion}", Config.APIVersion)
	return fmt.Sprintf("%s%s", vstsBaseURI, r.Replace(commentsURITemplate))
}

// GetReviewerURI constructs the URI for the reviewer API interations.
func GetReviewerURI(repositoryID string, pullRequestID string, reviewerID string) string {
	r := strings.NewReplacer(
		"{project}", Config.Project,
		"{repositoryId}", repositoryID,
		"{pullRequestId}", pullRequestID,
		"{reviewerId}", reviewerID,
		"{apiVersion}", Config.APIVersion)
	return fmt.Sprintf("%s%s", vstsBaseURI, r.Replace(reviewerURITemplate))
}

// GetPullRequestsURI constructs the URI for the pull requests API interactions.
func GetPullRequestsURI() string {
	r := strings.NewReplacer(
		"{project}", Config.Project,
		"{repositoryName}", Config.RepositoryName,
		"{apiVersion}", Config.APIVersion)

	return fmt.Sprintf("%s%s", vstsBaseURI, r.Replace(pullRequestsURITemplate))
}

func sendJSON(method string, url string, jsonData interface{}) error {
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

func getJSONResponse(url string, target interface{}) error {
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

// ContainsReviewBalancerComment checks if the passed in review has had a bot comment added.
func ContainsReviewBalancerComment(reviewSummary ReviewSummary) bool {
	url := GetCommentsURI(reviewSummary.RepositoryID, reviewSummary.ID)

	threads := new(CommentThreads)
	err := getJSONResponse(url, threads)

	// TODO: Make better error handling for this.
	if err != nil {
		log.Println(err)
		return false
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

// GetInprogressReviews gets all currently opened pull requests for a specific repository.
func GetInprogressReviews() []ReviewSummary {
	url := GetPullRequestsURI()

	pullRequests := new(PullRequests)
	err := getJSONResponse(url, pullRequests)
	if err != nil {
		log.Fatal(err)
	}

	reviewSummaries := make([]ReviewSummary, len(pullRequests.PullRequests))
	for index, pullRequest := range pullRequests.PullRequests {
		reviewSummaries[index] = NewReviewSummary(pullRequest)
	}
	return reviewSummaries
}

// AddRootComment adds a comment to the review passed in.
func AddRootComment(reviewSummary ReviewSummary, comment string) {
	thread := NewCommentThread(comment)

	url := GetCommentsURI(reviewSummary.RepositoryID, reviewSummary.ID)
	err := sendJSON("POST", url, thread)
	if err != nil {
		log.Fatal(err)
	}
}

// AddReviewers adds the passing in reviewers to the pull requests for the passed in review.
func AddReviewers(reviewSummary ReviewSummary, required []Reviewer, optional []Reviewer) {
	for _, reviewer := range append(required, optional...) {
		url := GetReviewerURI(reviewSummary.RepositoryID, reviewSummary.ID, reviewer.VisualStudioID)

		vote := NewDefaultReviewerVote()

		err := sendJSON("PUT", url, vote)
		if err != nil {
			log.Fatal(err)
		}
	}
}
