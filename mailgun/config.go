package mailgun

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mailgun/mailgun-go/v5"
	"log"
	"strings"
)

// Config struct holds API key
type Config struct {
	APIKey        string
	Region        string
	MailgunClient *mailgun.Client
}

// Client returns a new client for accessing mailgun.
func (c *Config) Client() (*Config, diag.Diagnostics) {

	log.Printf("[INFO] Mailgun Client configured ")

	return c, nil
}

// GetClient returns a client based on region.
func (c *Config) GetClient(Region string) (*mailgun.Client, error) {

	c.MailgunClient = mailgun.NewMailgun(c.APIKey)
	c.Region = Region
	c.ConfigureBaseUrl(Region)

	return c.MailgunClient, nil
}

func (c *Config) ConfigureBaseUrl(Region string) {
	if strings.ToLower(Region) == "eu" {
		_ = c.MailgunClient.SetAPIBase("https://api.eu.mailgun.net/v3")
	} else {
		_ = c.MailgunClient.SetAPIBase("https://api.mailgun.net/v3")
	}
}
