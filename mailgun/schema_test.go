package mailgun

import (
	"testing"
)

// schemaInvariants enforces a few low-level guarantees that are easy to
// regress when changing schemas. All resource-level invariants now live in
// internal/framework/; this file only covers the SDKv2 provider stub that
// remains for tf6muxserver compatibility.

func TestProviderSchema_ApiKeySensitive(t *testing.T) {
	field, ok := Provider().Schema["api_key"]
	if !ok {
		t.Fatal("provider must expose an api_key field")
	}
	if !field.Sensitive {
		t.Error("provider api_key must be marked Sensitive")
	}
	// api_key is Optional at the schema level (env var fallback handled in
	// providerConfigure) so the framework provider can declare an identical
	// schema and tf6muxserver can merge them without conflicts.
	if !field.Optional {
		t.Error("provider api_key must be Optional")
	}
}

func TestProvider_InternalValidate(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("provider failed internal validation: %s", err)
	}
}
