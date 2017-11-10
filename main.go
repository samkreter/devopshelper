package main

import (
	"fmt"
	"net/http"
	"github.com/spf13/viper"
	"net/url"
	"encoding/base64"
)

func setUpConfig(){
	viper.SetConfigName("config.dev") 
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil { 
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}


func getZillowSearchResult(address string, citystatezip string){
	u, err := url.Parse("https://msazure.visualstudio.com/")
	
	if err != nil{
		panic(err)
	}


	//test := "DefaultCollection/{project}/_apis/git/pullRequests?api-version={apiVersion}&reviewerId={reviewerId}"

	zwsId := viper.GetString("zwsId")

	q := u.Query()

	q.Add("zws-id",zwsId)
	q.Add("address", address)
	q.Add("citystatezip",citystatezip)

	u.RawQuery = q.Encode()
	fmt.Println(u)
}



func main(){

	setUpConfig()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization","Basic " + string(base64.StdEncoding.EncodeToString([]byte(viper.GetString("vstsToken")))))
	res, _ := client.Do(req)

	fmt.Println(res.Body)
}