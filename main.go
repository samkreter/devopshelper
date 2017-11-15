package main

import (
	"fmt"
	"github.com/samkreter/VSTSAutoReviewer/VSTS"
	"github.com/spf13/viper"

)

func setUpConfig(){
	viper.SetConfigName("config.dev") 
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil { 
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}


func main(){
	fmt.Println(VSTS.CommentsUriTemplate)
	panic("not implemented")

}