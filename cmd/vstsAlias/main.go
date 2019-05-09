package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	vsts "github.com/samkreter/vsts-goclient/client"
	"github.com/samkreter/devopshelper/pkg/config"
	"github.com/samkreter/devopshelper/pkg/utils"
)

var (
	conf    *config.Config
	inFile  string
	aliases string
	outFile string
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	flag.StringVar(&inFile, "inFile", "", "text file fo VSTS alias to convert to IDs")
	flag.StringVar(&inFile, "f", "", "text file fo VSTS alias to convert to IDs (shorthand)")
	flag.StringVar(&aliases, "aliases", "", "comma seperated list of aliases to convert")
	flag.StringVar(&aliases, "a", "", "comma seperated list of aliases to convert (shorthand)")
	flag.StringVar(&outFile, "outFile", "", "filepath to output the generated reviewers file to.")
	flag.StringVar(&outFile, "o", "", "filepath to output the generated reviewers file to (shorthand).")
	configFilePathPtr := flag.String("config-file", "", "filepath of the configuration file.")
	flag.Parse()

	if inFile == "" && aliases == "" {
		log.Fatal("Must provide either --aliases or --inputFile")
	}

	if *configFilePathPtr == "" {
		log.Fatal("Must suply a config location with --config-file")
	}

	var err error
	conf, err = config.LoadConfig(*configFilePathPtr)
	if err != nil {
		log.Fatal(err)
	}

	aliasMap := make(map[string]*Reviewer)
	if inFile != "" {
		err := addAliasFromFile(inFile, aliasMap)
		if err != nil {
			log.Fatal(err)
		}
	}
	if aliases != "" {
		addAliasesFromList(aliases, aliasMap)
	}

	vstsConfig := &vsts.Config{
		Token:      conf.Token,
		Username:   conf.Username,
		APIVersion: "5.0",
	}

	vstsClient, err := vsts.NewClient(vstsConfig)
	if err != nil {
		log.Fatal(err)
	}

	for alias := range aliasMap {
		devOpsIdentity, err := utils.GetDevOpsIdentity("sakreter", vstsClient.RestClient)
		if err != nil {
			log.Fatal(err)
		}

		aliasMap[alias] = &Reviewer{
			UniqueName: devOpsIdentity.Properties["Mail"].Value,
			Alias:      alias,
			ID:         devOpsIdentity.ID,
		}
	}

	if outFile != "" {
		err := generateReviewerFile(outFile, aliasMap)
		if err != nil {
			log.Fatal(err)
		}
	}

	for alias, reviewer := range aliasMap {
		if reviewer != nil {
			fmt.Printf("%s: %s\n", alias, reviewer.ID)
		} else {
			fmt.Printf("%s: Cloud not find ID, must have recently made a PR to get their ID.\n", alias)
		}
	}
}

func getAliasFromEmail(email string) string {
	splitEmail := strings.Split(email, "@")

	if len(splitEmail) == 2 {
		return splitEmail[0]
	}
	return ""
}

func addAliasesFromList(aliases string, aliasMap map[string]*Reviewer) {
	listAliases := strings.Split(aliases, ",")

	for _, alias := range listAliases {
		aliasMap[alias] = nil
	}
}

func addAliasFromFile(filePath string, aliasMap map[string]*Reviewer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		aliasMap[scanner.Text()] = nil
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// ReviewerGroups is a list of type ReviewerGroup
type ReviewerGroups []ReviewerGroup

// ReviewerPositions holds the current position information for the reviewers
type ReviewerPositions map[string]int

// ReviewerGroup holds the reviwers and metadata for a review group.
type ReviewerGroup struct {
	Group      string     `json:"group"`
	Required   bool       `json:"required"`
	Reviewers  []Reviewer `json:"reviewers"`
	CurrentPos int
}

// Reviewer is a vsts revier object
type Reviewer struct {
	UniqueName string `json:"uniqueName"`
	Alias      string `json:"alias"`
	ID         string `json:"id"`
}

func generateReviewerFile(outFile string, aliasMap map[string]*Reviewer) error {
	reviewers := make([]Reviewer, 0, len(aliasMap))

	for _, reviewer := range aliasMap {
		if reviewer.ID != "" {
			reviewers = append(reviewers, *reviewer)
		}
	}

	b, err := json.MarshalIndent(reviewers, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outFile, b, 0644)
	if err != nil {
		return err
	}

	return nil
}
