package main

import (
	"fmt"
	"github.com/samkreter/VSTSAutoReviewer/VSTS"
)


func main(){
	fmt.Println(VSTS.GetPullRequestsUri())
}