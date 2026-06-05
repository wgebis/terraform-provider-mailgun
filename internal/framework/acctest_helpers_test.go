package framework_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/joho/godotenv"
	"github.com/mailgun/mailgun-go/v5"

	"github.com/wgebis/terraform-provider-mailgun/internal/framework"
	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

// TestMain loads a local .env file (if present) before running the test suite
// so acceptance tests can pick up MAILGUN_API_KEY without exporting it in the
// shell. Existing environment variables always take precedence.
func TestMain(m *testing.M) {
	loadDotEnv()
	os.Exit(m.Run())
}

// loadDotEnv searches the current directory and its parents for a .env file
// and loads it. Missing files are ignored; this is a test-only convenience.
func loadDotEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			_ = godotenv.Load(candidate)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

// protoV6Providers returns the provider factory used by acceptance tests in
// this package.
func protoV6Providers() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"mailgun": func() (tfprotov6.ProviderServer, error) {
			return framework.NewProviderServer()
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
