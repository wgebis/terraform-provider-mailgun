package mailgun

import (
	"fmt"
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
	err := c.ConfigureBaseUrl(Region)
	if err != nil {
		return nil, fmt.Errorf("Error configuring base URL: %s", err)
	}

	return c.MailgunClient, nil
}

func (c *Config) ConfigureBaseUrl(Region string) error {
	var baseUrl string
	if strings.ToLower(Region) == "eu" {
		baseUrl = "https://api.eu.mailgun.net"
	} else {
		baseUrl = "https://api.mailgun.net"
	}
	
	log.Printf("[DEBUG] Setting Mailgun API base URL for region %s: %s", Region, baseUrl)
	err := c.MailgunClient.SetAPIBase(baseUrl)
	if err != nil {
		return fmt.Errorf("Error setting API base URL to %s: %s", baseUrl, err)
	}
	
	log.Printf("[DEBUG] Successfully configured Mailgun client for region %s", Region)
	return nil
}
