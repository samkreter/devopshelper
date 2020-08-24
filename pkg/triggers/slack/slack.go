package slack

import (
	"fmt"
	"strings"

	"github.com/samkreter/devopshelper/pkg/types"
)

const (
	triggerName     = "slackTrigger"
	messageTemplate = "%s, you have a PR to review: %s"
)

type ReviwerTrigger func([]*types.Reviewer, string) error

// SlackConfig configuration for the slack reviwer trigger
type SlackConfig struct {
	Token        string            `json:"token"`
	Channel      string            `json:"channel"`
	AliasConvert map[string]string `json:"aliasConvert"`
}

// NewSlackTrigger creates a new slack trigger to mention reviewers
func NewSlackTrigger(config *SlackConfig) (ReviwerTrigger, error) {
	c := NewClient(config.Token)

	trigger := func(reviewers []*types.Reviewer, pullRequestURL string) error {
		slackUsers := make([]string, 0, len(reviewers))

		for _, reviewer := range reviewers {

			// Note: If there are no convertion for the reviewer alias, default to the alias
			slackUser, ok := config.AliasConvert[reviewer.Alias]
			if !ok {
				slackUser = reviewer.Alias
			}

			slackUsers = append(slackUsers, GetMention(slackUser))
		}

		msg := fmt.Sprintf(messageTemplate, strings.Join(slackUsers, " "), pullRequestURL)

		if err := c.SendMessage(msg, config.Channel); err != nil {
			return err
		}

		return nil
	}

	return trigger, nil
}
