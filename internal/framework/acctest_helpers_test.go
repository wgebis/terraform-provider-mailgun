package framework_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mailgun/mailgun-go/v5"

	"github.com/wgebis/terraform-provider-mailgun/internal/framework"
	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

// protoV6Providers returns the muxed provider factory used by acceptance
// tests in this package.
func protoV6Providers() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"mailgun": func() (tfprotov6.ProviderServer, error) {
			return framework.MuxedProviderServer(context.Background())
		},
	}
}

// testAccPreCheck mirrors the legacy SDKv2 helper.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("MAILGUN_API_KEY"); v == "" {
		t.Fatal("MAILGUN_API_KEY must be set for acceptance tests")
	}
}

// mailgunClientFromAttrs builds a Mailgun client from the resource state's
// region attribute and the MAILGUN_API_KEY env var. Replaces the legacy
// pattern of pulling Meta() off the SDKv2 provider.
func mailgunClientFromAttrs(attrs map[string]string) (*mailgun.Client, error) {
	cfg := &mailgunpkg.Config{APIKey: os.Getenv("MAILGUN_API_KEY")}
	region := attrs["region"]
	if region == "" {
		region = "us"
	}
	return cfg.GetClient(region)
}

// testAccCheckAnyAttrMatches asserts that at least one element of a list/set
// attribute on the named resource has its `subkey` matching the given regex.
func testAccCheckAnyAttrMatches(resourceName, listAttr, subkey string, re *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		prefix := listAttr + "."
		suffix := "." + subkey
		for k, v := range rs.Primary.Attributes {
			if !strings.HasPrefix(k, prefix) || !strings.HasSuffix(k, suffix) {
				continue
			}
			if re.MatchString(v) {
				return nil
			}
		}
		return fmt.Errorf("no element of %s.*.%s matched %q on %s",
			listAttr, subkey, re.String(), resourceName)
	}
}
