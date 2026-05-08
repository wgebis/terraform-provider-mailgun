package mailgun_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	mailgunclient "github.com/mailgun/mailgun-go/v5"

	"github.com/wgebis/terraform-provider-mailgun/internal/framework"
	"github.com/wgebis/terraform-provider-mailgun/mailgun"
)

// protoV6Providers returns the muxed provider factory used by acceptance
// tests in this package. Combines the legacy SDKv2 provider with the new
// terraform-plugin-framework provider.
func protoV6Providers() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"mailgun": func() (tfprotov6.ProviderServer, error) {
			return framework.MuxedProviderServer(context.Background())
		},
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("MAILGUN_API_KEY"); v == "" {
		t.Fatal("MAILGUN_API_KEY must be set for acceptance tests")
	}
}

// mailgunClientFor builds a Mailgun client from MAILGUN_API_KEY env var for
// the given region. Replaces the legacy `testAccProvider.Meta().(*Config)`
// pattern after the move to an external `mailgun_test` package.
func mailgunClientFor(region string) (*mailgunclient.Client, error) {
	if region == "" {
		region = "us"
	}
	cfg := &mailgun.Config{APIKey: os.Getenv("MAILGUN_API_KEY")}
	return cfg.GetClient(region)
}
