package mailgun

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// schemaInvariants enforces a few low-level guarantees that are easy to
// regress when changing schemas: required vs optional consistency, and
// that secrets (passwords, API keys) stay marked Sensitive.

func TestProviderSchema_ApiKeySensitive(t *testing.T) {
	field, ok := Provider().Schema["api_key"]
	if !ok {
		t.Fatal("provider must expose an api_key field")
	}
	if !field.Sensitive {
		t.Error("provider api_key must be marked Sensitive")
	}
	if !field.Required {
		t.Error("provider api_key must be Required")
	}
}

func TestCredentialSchema_PasswordSensitive(t *testing.T) {
	pwd := resourceMailgunCredential().Schema["password"]
	if !pwd.Sensitive {
		t.Error("credential password must be marked Sensitive")
	}
}

func TestApiKeySchema_SecretSensitiveAndComputed(t *testing.T) {
	secret := resourceMailgunApiKey().Schema["secret"]
	if !secret.Sensitive {
		t.Error("api_key.secret must be marked Sensitive")
	}
	if !secret.Computed {
		t.Error("api_key.secret must be Computed")
	}
}

func TestRegionDefault(t *testing.T) {
	resources := map[string]*schema.Resource{
		"mailgun_route":             resourceMailgunRoute(),
		"mailgun_webhook":           resourceMailgunWebhook(),
		"mailgun_domain_credential": resourceMailgunCredential(),
		"mailgun_domain":            resourceMailgunDomain(),
		"mailgun_api_key":           resourceMailgunApiKey(),
	}

	for name, r := range resources {
		t.Run(name, func(t *testing.T) {
			region, ok := r.Schema["region"]
			if !ok {
				t.Fatalf("%s: missing region attribute", name)
			}
			if !region.Optional {
				t.Errorf("%s: region must be Optional", name)
			}
			if !region.ForceNew {
				t.Errorf("%s: region must be ForceNew", name)
			}
			if region.Default != "us" {
				t.Errorf("%s: region default = %v, want \"us\"", name, region.Default)
			}
		})
	}
}

func TestProvider_InternalValidate(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider failed internal validation: %s", err)
	}
}
