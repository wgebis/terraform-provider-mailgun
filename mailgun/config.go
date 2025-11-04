package mailgun

import (
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mailgun/mailgun-go/v5"
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
		_ = c.MailgunClient.SetAPIBase(mailgun.APIBaseEU)
	} else {
		_ = c.MailgunClient.SetAPIBase(mailgun.APIBase)
	}
}
