package VSTS

import (
	"fmt"
	"net/http"
	"github.com/spf13/viper"
	"net/url"
	"io/ioutil"
	"strings"
)

var PullRequestsUriTemplate string = "DefaultCollection/{project}/_apis/git/pullRequests?api-version={apiVersion}&reviewerId={reviewerId}"
var CommentsUriTemplate string = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/threads?api-version={apiVersion}"
var eviewerUriTemplate string = "DefaultCollection/_apis/git/repositories/{repositoryId}/pullRequests/{pullRequestId}/reviewers/{reviewerId}?api-version={apiVersion}"
var vstsBaseUri string = "https://msazure.visualstudio.com/"

func setUpConfig(){
	viper.SetConfigName("config.dev") 
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil { 
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}


func gettest(){
	u, err := url.Parse(vstsBaseUri)
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
	setUpConfig()
	client := &http.Client{}
	url := "https://msazure.VisualStudio.com/DefaultCollection/_apis/projects?api-version=2.0"
	req, _ := http.NewRequest("GET", url, nil)
	username := viper.GetString("vstsUsername")
	token := viper.GetString("vstsToken")
	fmt.Println(username,token)
	req.SetBasicAuth(username, token)

	res, _ := client.Do(req)
	bodyText, err := ioutil.ReadAll(res.Body)
	if err != nil{
		panic(err)
	}
	fmt.Println(string(bodyText))

}