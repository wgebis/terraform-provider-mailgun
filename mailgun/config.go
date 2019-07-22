package mailgun

import (
	"log"

	mailgun "github.com/mailgun/mailgun-go/v3"
)

type Config struct {
	APIKey string
}

// Client() returns a new client for accessing mailgun.
//
func (c *Config) Client() (*mailgun.MailgunImpl, error) {

	client := mailgun.NewMailgun("", c.APIKey)

	log.Printf("[INFO] Mailgun Client configured ")

	return client, nil
}
