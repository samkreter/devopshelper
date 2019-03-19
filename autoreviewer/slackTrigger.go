package autoreviewer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/samkreter/vstsautoreviewer/slack"
)

const (
	triggerName     = "slackTrigger"
	messageTemplate = "%s, you have a PR to review: %s"
)

// SlackConfig configuration for the slack reviwer trigger
type SlackConfig struct {
	Token        string            `json:"token"`
	Channel      string            `json:"channel"`
	AliasConvert map[string]string `json:"aliasConvert"`
}

// NewSlackTrigger creates a new slack trigger to mention reviewers
func NewSlackTrigger(configPath string) (ReviwerTrigger, error) {
	config, err := getConfig(configPath)
	if err != nil {
		return nil, err
	}

	c := slack.NewClient(config.Token)

	trigger := func(reviewers []Reviewer, pullRequestURL string) error {
		slackUsers := make([]string, 0, len(reviewers))

		for _, reviewer := range reviewers {

			// Note: If there are no convertion for the reviewer alias, default to the alias
			slackUser, ok := config.AliasConvert[reviewer.Alias]
			if !ok {
				slackUser = reviewer.Alias
			}

			slackUsers = append(slackUsers, slack.GetMention(slackUser))
		}

		msg := fmt.Sprintf(messageTemplate, strings.Join(slackUsers, " "), pullRequestURL)

		if err := c.SendMessage(msg, config.Channel); err != nil {
			return err
		}

		return nil
	}

	return trigger, nil
}

func getConfig(filePath string) (*SlackConfig, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("File %s does not exist", filePath)
	}

	config := &SlackConfig{}
	err = json.Unmarshal(b, config)
	if err != nil {
		return nil, err
	}

	if config.Token == "" {
		return nil, fmt.Errorf("%s: missing slack token in configuration", triggerName)
	}

	return config, nil
}
