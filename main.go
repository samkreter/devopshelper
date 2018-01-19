package main

import (
	"log"
	"time"

	"github.com/samkreter/VSTSAutoReviewer/cmd"
)

func main() {

	timer := time.NewTicker(time.Second * 1)

	for _ = range timer.C {
		if err := cmd.Run(); err != nil {
			log.Println("Error for main run", err)
		}
		log.Println("Running Review cycle...")
	}
}
