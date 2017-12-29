package main

import (
	"log"

	"github.com/samkreter/VSTSAutoReviewer/cmd"
)

func main() {
	if err := cmd.RunTest(); err != nil {
		log.Fatal(err)
	}
}
