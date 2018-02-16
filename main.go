package main

import (
	"log"
	"time"

	"github.com/samkreter/VSTSAutoReviewer/cmd"
)

func main() {

	log.Println("Starting Reviewer...")

	if err := cmd.Run(); err != nil {
		log.Println("Error for main run", err)
	}
	log.Println(time.Now(), ": Finished Balancing Cycle")

}
