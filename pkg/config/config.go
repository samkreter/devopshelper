package config

import (
	"encoding/json"
	"os"
)

var (
	defaultReviewerFile = "/configs/reviewers.json"
	defaultStatusFile   = "/configs/currentStatus.json"
)

// Config holds the configuration from the config file
type Config struct {
	Token           string            `json:"token"`
	Username        string            `json:"username"`
	APIVersion      string            `json:"apiVersion"`
	BotMaker        string            `json:"botMaker"`
	RepositoryInfos []*RepositoryInfo `json:"repositoryInfos"`
	Instance        string            `json:"instance"`
}

// RepositoryInfo information describing each repository to review
type RepositoryInfo struct {
	ProjectName    string `json:"projectName"`
	RepositoryName string `json:"repositoryName"`
	ReviewerFile   string `json:"reviewerFile"`
	StatusFile     string `json:"reviewerStatusFile"`
}

// LoadConfig loads the reviewer configuration from a json file
func LoadConfig(configFilePath string) (*Config, error) {
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}

	defer configFile.Close()

	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
