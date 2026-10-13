package mailgun

import (
	"net/http"
	"strings"
	"sync"

	"github.com/mailgun/mailgun-go/v5"
)

// Config struct holds API key and optional request pacing settings.
type Config struct {
	APIKey string

	// RequestsPerSecond caps outgoing API requests for the whole provider
	// instance. 0 disables pacing (upstream behaviour).
	RequestsPerSecond float64
	// MaxRetries is how many times a 429 response is retried. 0 disables.
	MaxRetries int

	httpClientOnce sync.Once
	httpClient     *http.Client
}

// sharedHTTPClient returns one http.Client for the provider instance so the
// rate limiter is shared by every resource, not recreated per operation.
func (c *Config) sharedHTTPClient() *http.Client {
	c.httpClientOnce.Do(func() {
		if c.RequestsPerSecond <= 0 && c.MaxRetries <= 0 {
			c.httpClient = http.DefaultClient
			return
		}
		c.httpClient = &http.Client{
			Transport: newRateLimitedTransport(c.RequestsPerSecond, c.MaxRetries),
		}
	})
	return c.httpClient
}

// GetClient returns a fresh Mailgun client for the given region. A new client
// is constructed on every call so concurrent operations targeting different
// regions do not race on a shared client instance. The underlying HTTP client
// (and therefore the rate limiter) is shared across all of them.
func (c *Config) GetClient(region string) (*mailgun.Client, error) {

	client := mailgun.NewMailgun(c.APIKey)
	client.SetHTTPClient(c.sharedHTTPClient())
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
