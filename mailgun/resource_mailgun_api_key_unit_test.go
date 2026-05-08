package mailgun

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// TestApplyAPIKey_PreservesSecret is a regression test for issue #73:
// Mailgun returns the secret only on creation. Subsequent reads must not
// nullify the locally-stored secret.
func TestApplyAPIKey_PreservesSecret(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceMailgunApiKey().Schema, map[string]interface{}{})
	d.SetId("k1")
	_ = d.Set("secret", "previously-stored-secret")

	applyAPIKey(d, mtypes.APIKey{
		ID:          "k1",
		Description: "ci key",
		Kind:        "user",
		Role:        "admin",
		// Secret intentionally empty to mimic the list endpoint response.
	})

	if got := d.Get("secret").(string); got != "previously-stored-secret" {
		t.Errorf("secret was nullified, got %q", got)
	}
	if got := d.Get("description").(string); got != "ci key" {
		t.Errorf("description not synced, got %q", got)
	}
	if got := d.Get("kind").(string); got != "user" {
		t.Errorf("kind not synced, got %q", got)
	}
	if got := d.Get("role").(string); got != "admin" {
		t.Errorf("role not synced, got %q", got)
	}
}

func TestApplyAPIKey_OverwritesSecretWhenProvided(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceMailgunApiKey().Schema, map[string]interface{}{})
	d.SetId("k1")
	_ = d.Set("secret", "old")

	applyAPIKey(d, mtypes.APIKey{ID: "k1", Secret: "new"})

	if got := d.Get("secret").(string); got != "new" {
		t.Errorf("secret should be overwritten when API returns a value, got %q", got)
	}
}

func TestApplyAPIKey_SyncsDisabledFlags(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceMailgunApiKey().Schema, map[string]interface{}{})
	d.SetId("k1")

	applyAPIKey(d, mtypes.APIKey{
		ID:             "k1",
		IsDisabled:     true,
		DisabledReason: "rotated",
	})

	if !d.Get("is_disabled").(bool) {
		t.Error("is_disabled not synced")
	}
	if got := d.Get("disabled_reason").(string); got != "rotated" {
		t.Errorf("disabled_reason = %q, want \"rotated\"", got)
	}
}
