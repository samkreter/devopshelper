package VSTS

import (
	"fmt"
	"net/http"
	"github.com/spf13/viper"
	"net/url"
	"io/ioutil"
	"strings"
)

type config struct {
	VstsToken 		string	`json:"vstsToken"`
	VstsProject		string	`json:"vstsProject"`
	VstsUsername	string	`json:"vstsUsername"`
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

//TODO: set up correct replacements
func GetCommentsUri() string{
	r := strings.NewReplacer(	"{project}", 	viper.GetString("vstsProject"),
							 	"{apiVersion}", ApiVersion)
	return fmt.Sprintf("%s%s",VstsBaseUri,r.Replace(ReviewerUriTemplate))
}

//TODO set up correct replacements
func GetReviewerUri() string{
	r := strings.NewReplacer(	"{repositoryId}", 	conf.VstsProject,
								"{pullRequestId", 	"pullRequestId",
								"{reviewerId}",		"reviewerId",
							 	"{apiVersion}", 	ApiVersion)
	return fmt.Sprintf("%s%s",VstsBaseUri,r.Replace(ReviewerUriTemplate))
}

func GetPullRequestsUri() string{
	r := strings.NewReplacer("{project}", viper.GetString("vstsProject"),
							"{apiVersion}", ApiVersion)
	return fmt.Sprintf("%s%s",VstsBaseUri,r.Replace(PullRequestsUriTemplate))
}

// func GetInprogressReviews(){

// 	client := &http.Client{}
// 	url := GetPullRequestsUri()
// 	req, _ := http.NewRequest("GET", url, nil)
// 	req.SetBasicAuth(username, token)

// 	res, _ := client.Do(req)


// 	var pullRequests = await response.Content
// 		.ReadAsAsync<VisualStudioPullRequests>()
// 		.ConfigureAwait(continueOnCapturedContext: false);

//    return pullRequests.PullRequests
// 		.Select(pr => ReviewSummary.GetReviewSummary(pr))
// 		.ToArray();
// }

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