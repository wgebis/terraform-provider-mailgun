package mailgun

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mailgun/mailgun-go/v3"
	"log"
	"strings"
)

// Config struct holds API key
//
type Config struct {
	APIKey string
	//Region        string
	//Domain        string
	MailgunClient *mailgun.MailgunImpl
}

// Client returns a new client for accessing mailgun.
//
func (c *Config) Client() (*Config, diag.Diagnostics) {

	log.Printf("[INFO] Mailgun Client configured ")

	return c, nil
}

// Client returns a client based on region.
//

func (c *Config) GetClientForDomain(Region string, Domain string) (*mailgun.MailgunImpl, error) {

	c.MailgunClient = mailgun.NewMailgun(Domain, c.APIKey)
	c.ConfigureBaseUrl(Region)

	return c.MailgunClient, nil
}

func (c *Config) GetClient(Region string) (*mailgun.MailgunImpl, error) {

	mc, _ := c.GetClientForDomain(Region, "")

	return mc, nil
}

func (c *Config) ConfigureBaseUrl(Region string) {
	if strings.ToLower(Region) == "eu" {
		c.MailgunClient.SetAPIBase("https://api.eu.mailgun.net/v3")
	} else {
		c.MailgunClient.SetAPIBase("https://api.mailgun.net/v3")
	}
}
