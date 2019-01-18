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

	"github.com/antihax/optional"

	vstsObj "github.com/samkreter/vsts-goclient/api/git"
	vsts "github.com/samkreter/vsts-goclient/client"
)

var (
	config  Config
	inFile  string
	aliases string
	outFile string
)

// Config holds the configuration from the config file
type Config struct {
	Token          string `json:"token"`
	Username       string `json:"username"`
	APIVersion     string `json:"apiVersion"`
	RepositoryName string `json:"repositoryName"`
	Project        string `json:"project"`
	Instance       string `json:"instance"`
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configFilePath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		log.Fatal("CONFIG_PATH not set")
	}

	configFile, err := os.Open(configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	defer configFile.Close()

	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.StringVar(&inFile, "inFile", "", "text file fo VSTS alias to convert to IDs")
	flag.StringVar(&inFile, "f", "", "text file fo VSTS alias to convert to IDs (shorthand)")
	flag.StringVar(&aliases, "aliases", "", "comma seperated list of aliases to convert")
	flag.StringVar(&aliases, "a", "", "comma seperated list of aliases to convert (shorthand)")
	flag.StringVar(&outFile, "outFile", "", "filepath to output the generated reviewers file to.")
	flag.StringVar(&outFile, "o", "", "filepath to output the generated reviewers file to (shorthand).")
	flag.Parse()

	if inFile == "" && aliases == "" {
		log.Fatal("Must provide either --aliases or --inputFile")
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
		Token:          config.Token,
		Username:       config.Username,
		APIVersion:     config.APIVersion,
		RepositoryName: config.RepositoryName,
		Project:        config.Project,
		Instance:       config.Instance,
	}

	vstsClient, err := vsts.NewClient(vstsConfig)
	if err != nil {
		log.Fatal(err)
	}

	getOpts := &vstsObj.GetPullRequestsOpts{
		SearchCriteriaStatus: optional.NewString("all"),
		Top:                  optional.NewInt32(1000),
	}

	pullRequests, err := vstsClient.GetPullRequests(getOpts)
	if err != nil {
		log.Fatalf("get pull requests error: %v", err)
	}

	for _, pullRequest := range pullRequests {
		alias := getAliasFromEmail(pullRequest.CreatedBy.UniqueName)

		if _, ok := aliasMap[alias]; ok {
			aliasMap[alias] = &Reviewer{
				UniqueName: pullRequest.CreatedBy.UniqueName,
				Alias:      alias,
				ID:         pullRequest.CreatedBy.ID,
			}
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
