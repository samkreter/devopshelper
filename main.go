package main

import (
	"log"

	"github.com/samkreter/VSTSAutoReviewer/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
