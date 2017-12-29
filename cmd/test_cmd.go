package cmd

import (
	"fmt"

	"github.com/samkreter/VSTSAutoReviewer/vsts"
)

func RunTest() error {
	fmt.Println(vsts.Config.APIVersion)

	return nil
}
