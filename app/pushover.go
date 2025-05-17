package app

import (
	"errors"

	"github.com/gregdel/pushover"
)

type NotifyClient struct {
	app       *pushover.Pushover
	recipient *pushover.Recipient
}

func NewNotifyClient(appToken, recipientToken string) (*NotifyClient, error) {
	if appToken == "" {
		return nil, errors.New("missing required app_token")
	}
	if recipientToken == "" {
		return nil, errors.New("missing required recipient_token")
	}

	return &NotifyClient{
		app:       pushover.New(appToken),
		recipient: pushover.NewRecipient(recipientToken),
	}, nil
}

func (c *NotifyClient) Send(title, message string) error {
	msg := pushover.NewMessageWithTitle(message, title)
	_, err := c.app.SendMessage(msg, c.recipient)
	return err
}
