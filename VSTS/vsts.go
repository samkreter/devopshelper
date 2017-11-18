package VSTS

import (
	"log"
	"fmt"
	"net/http"
	"github.com/spf13/viper"
	"net/url"
	"io/ioutil"
	"strings"
	"encoding/json"
	"time"
)

type config struct {
	VstsToken 			string	`json:"vstsToken"`
	VstsProject			string	`json:"vstsProject"`
	VstsUsername		string	`json:"vstsUsername"`
	VstsRepositoryId	string	`json:"repositoryId"`
	VstsArmReviewerId	string 	`json:"vstsArmReviewerId"`
}

var (
	conf *config
	PullRequestsUriTemplate string = "DefaultCollection/{project}/_apis/git/pullRequests?api-version={apiVersion}&reviewerId={reviewerId}"
	CommentsUriTemplate string = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/threads?api-version={apiVersion}"
	ReviewerUriTemplate string = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/reviewers/{reviewerId}?api-version={apiVersion}"
	VstsBaseUri string = "https://msazure.visualstudio.com/"
	ApiVersion string = "3.0"
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

func init(){
	conf = getConf()
	fmt.Println(conf)
}

func GetCommentsUri(pullRequestId string, repositoryId string) string{
	r := strings.NewReplacer(	"{repositoryId}", 	repositoryId,
								"{pullRequestId", 	pullRequestId,
							 	"{apiVersion}", 	ApiVersion)
	return fmt.Sprintf("%s%s",VstsBaseUri,r.Replace(ReviewerUriTemplate))
}

func GetReviewerUri(repositoryId string, pullRequestId string, reviewerId string) string{
	r := strings.NewReplacer(	"{repositoryId}", 	repositoryId,
								"{pullRequestId", 	pullRequestId,
								"{reviewerId}",		reviewerId,
							 	"{apiVersion}", 	ApiVersion)
	return fmt.Sprintf("%s%s",VstsBaseUri,r.Replace(ReviewerUriTemplate))
}

func GetPullRequestsUri() string{
	r := strings.NewReplacer(	"{project}", 		conf.VstsProject,
							 	"{reviewerId}",		conf.VstsArmReviewerId,
								"{apiVersion}", 	ApiVersion)
	return fmt.Sprintf("%s%s",VstsBaseUri,r.Replace(PullRequestsUriTemplate))
}

func GetJsonResponse(url string, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(conf.VstsUsername, conf.VstsToken)

	res, err := client.Do(req)
	if err != nil {
        return err
    }
	
	defer res.Body.Close()

    return json.NewDecoder(res.Body).Decode(target)
}

func GetInprogressReviews() []ReviewSummary{
	url := GetPullRequestsUri()

	pullRequests := new(VstsPullRequests)
	err := GetJsonResponse(url, pullRequests)
	if err != nil{
		log.Fatal(err)
	}

	reviewSummaries := make([]ReviewSummary,len(pullRequests.PullRequests))
	for index, pullRequest := range pullRequests.PullRequests{
		reviewSummary := new(ReviewSummary)
		reviewSummaries[index] = reviewSummary.GetReviewSummary(pullRequest)
	}
	return reviewSummaries
}

func gettest(){
	u, err := url.Parse(VstsBaseUri)
	if err != nil{
		panic(err)
	}

	r := strings.NewReplacer(	"{project}", viper.GetString("vstsProject"),
            					"{apiVersion}", viper.GetString("vstsApiVersion"))


	result := r.Replace(PullRequestsUriTemplate)
	fmt.Println(result)

	q := u.Query()

	// q.Add("address", address)
	// q.Add("citystatezip",citystatezip)

	u.RawQuery = q.Encode()
	fmt.Println(u)
}

func initalize(){
	client := &http.Client{}
	url := "https://msazure.VisualStudio.com/DefaultCollection/_apis/projects?api-version=2.0"
	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(conf.VstsUsername, conf.VstsToken)

	res, _ := client.Do(req)
	bodyText, err := ioutil.ReadAll(res.Body)
	if err != nil{
		panic(err)
	}
	fmt.Println(string(bodyText))

}