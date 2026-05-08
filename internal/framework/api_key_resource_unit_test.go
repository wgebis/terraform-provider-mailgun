package framework

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// TestApplyAPIKeyToModel_PreservesSecret is a regression test for issue #73:
// Mailgun returns the secret only on creation. Subsequent reads must not
// nullify the locally-stored secret.
func TestApplyAPIKeyToModel_PreservesSecret(t *testing.T) {
	m := &apiKeyResourceModel{
		ID:     types.StringValue("k1"),
		Secret: types.StringValue("previously-stored-secret"),
	}

	applyAPIKeyToModel(m, mtypes.APIKey{
		ID:          "k1",
		Description: "ci key",
		Kind:        "user",
		Role:        "admin",
		// Secret intentionally empty to mimic the list endpoint response.
	})

	if got := m.Secret.ValueString(); got != "previously-stored-secret" {
		t.Errorf("secret was nullified, got %q", got)
	}
	if got := m.Description.ValueString(); got != "ci key" {
		t.Errorf("description not synced, got %q", got)
	}
	if got := m.Kind.ValueString(); got != "user" {
		t.Errorf("kind not synced, got %q", got)
	}
	if got := m.Role.ValueString(); got != "admin" {
		t.Errorf("role not synced, got %q", got)
	}
}

func TestApplyAPIKeyToModel_OverwritesSecretWhenProvided(t *testing.T) {
	m := &apiKeyResourceModel{
		ID:     types.StringValue("k1"),
		Secret: types.StringValue("old"),
	}

	applyAPIKeyToModel(m, mtypes.APIKey{ID: "k1", Secret: "new"})

	if got := m.Secret.ValueString(); got != "new" {
		t.Errorf("secret should be overwritten when API returns a value, got %q", got)
	}
}

func TestApplyAPIKeyToModel_SyncsDisabledFlags(t *testing.T) {
	m := &apiKeyResourceModel{ID: types.StringValue("k1")}

	applyAPIKeyToModel(m, mtypes.APIKey{
		ID:             "k1",
		IsDisabled:     true,
		DisabledReason: "rotated",
	})

	if !m.IsDisabled.ValueBool() {
		t.Error("is_disabled not synced")
	}
	if got := m.DisabledReason.ValueString(); got != "rotated" {
		t.Errorf("disabled_reason = %q, want \"rotated\"", got)
	}
}
