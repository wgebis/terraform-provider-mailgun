package framework

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func credentialSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	(&credentialResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	return resp
}

// credentialConfig builds a config with every attribute null except the ones
// given in set.
func credentialConfig(t *testing.T, set map[string]tftypes.Value) tfsdk.Config {
	t.Helper()
	s := credentialSchema(t).Schema
	objType := s.Type().TerraformType(context.Background()).(tftypes.Object)
	vals := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		vals[name] = tftypes.NewValue(typ, nil)
	}
	for name, v := range set {
		vals[name] = v
	}
	return tfsdk.Config{Schema: s, Raw: tftypes.NewValue(objType, vals)}
}

func TestCredentialSchema_PasswordWOIsWriteOnly(t *testing.T) {
	s := credentialSchema(t).Schema
	if diags := s.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("schema implementation invalid: %v", diags)
	}
	if !s.Attributes["password_wo"].IsWriteOnly() {
		t.Error("password_wo must be write-only")
	}
	if s.Attributes["password"].IsRequired() {
		t.Error("password must be optional so password_wo can replace it")
	}
}

// Write-only values are null in the plan, so Create and Update must read
// password_wo from the config.
func TestCredentialPassword_PrefersWriteOnlyFromConfig(t *testing.T) {
	cfg := credentialConfig(t, map[string]tftypes.Value{
		"password_wo":         tftypes.NewValue(tftypes.String, "from-config"),
		"password_wo_version": tftypes.NewValue(tftypes.Number, 1),
	})
	plan := credentialResourceModel{Password: types.StringNull(), PasswordWO: types.StringNull()}

	got, diags := credentialPassword(context.Background(), cfg, plan)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	if got != "from-config" {
		t.Errorf("password = %q, want the password_wo config value", got)
	}
}

func TestCredentialPassword_FallsBackToPassword(t *testing.T) {
	cfg := credentialConfig(t, map[string]tftypes.Value{
		"password": tftypes.NewValue(tftypes.String, "plain"),
	})
	plan := credentialResourceModel{Password: types.StringValue("plain"), PasswordWO: types.StringNull()}

	got, diags := credentialPassword(context.Background(), cfg, plan)
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	if got != "plain" {
		t.Errorf("password = %q, want %q", got, "plain")
	}
}
