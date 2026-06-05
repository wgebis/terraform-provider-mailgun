package mailgun

import (
	"strings"

	"github.com/mailgun/mailgun-go/v5"
)

// Config struct holds API key
type Config struct {
	APIKey string
}

// GetClient returns a fresh Mailgun client for the given region. A new client
// is constructed on every call so concurrent operations targeting different
// regions do not race on a shared client instance.
func (c *Config) GetClient(region string) (*mailgun.Client, error) {

	client := mailgun.NewMailgun(c.APIKey)
	configureBaseUrl(client, region)

	return client, nil
}

func configureBaseUrl(client *mailgun.Client, region string) {
	if strings.ToLower(region) == "eu" {
		_ = client.SetAPIBase(mailgun.APIBaseEU)
	} else {
		_ = client.SetAPIBase(mailgun.APIBase)
	}
}
