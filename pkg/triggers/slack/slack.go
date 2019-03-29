package slack

import (
	"fmt"
	"log"

	"github.com/nlopes/slack"
)

// Client enables interactiosn with the slack API
type Client struct {
	client *slack.Client
}

// NewClient creates a new slack API client
func NewClient(slackToken string) *Client {
	api := slack.New(slackToken)
	return &Client{
		client: api,
	}
}

// SendMessage send a message to a user
func (s *Client) SendMessage(text, channel string) error {
	_, _, respText, err := s.client.SendMessage(channel, slack.MsgOptionText(text, false))
	if err != nil {
		return err
	}

	log.Println("Slack Respsonse text: ", respText)

	return nil
}

// GetUsers gets all users accessable by the Auth Token
func (s *Client) GetUsers() ([]slack.User, error) {
	return s.client.GetUsers()
}

// GetMention returns the text to mention the name
func GetMention(name string) string {
	return fmt.Sprintf("<@%s>", name)
}
