package mailgun

import (
	"log"
	"strings"

	"github.com/mailgun/mailgun-go/v3"
)

// Config struct holds API key
//
type Config struct {
	APIKey   string
	USClient *mailgun.MailgunImpl
	EUClient *mailgun.MailgunImpl
}

// Client returns a new client for accessing mailgun.
//
func (c *Config) Client() (*Config, error) {

	c.USClient = mailgun.NewMailgun("", c.APIKey)
	c.USClient.SetAPIBase("https://api.mailgun.net/v3")
	c.EUClient = mailgun.NewMailgun("", c.APIKey)
	c.EUClient.SetAPIBase("https://api.eu.mailgun.net/v3")

	log.Printf("[INFO] Mailgun Client configured ")

	return c, nil
}

// Client returns a client based on region.
//
func (c *Config) GetClient(Region string) (*mailgun.MailgunImpl, error) {

	if strings.ToLower(Region) == "eu" {
		return c.EUClient, nil
	} else if strings.ToLower(Region) == "us" {
		return c.USClient, nil
	} else {
		// fallback to default region
		return c.USClient, nil
	}
}
