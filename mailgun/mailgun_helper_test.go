package mailgun

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mailgun/mailgun-go/v5"
)

func TestStringHashcode_Deterministic(t *testing.T) {
	a := stringHashcode("foo.example.com")
	b := stringHashcode("foo.example.com")
	if a != b {
		t.Fatalf("stringHashcode not deterministic: %d != %d", a, b)
	}
	if a < 0 {
		t.Fatalf("stringHashcode returned negative value: %d", a)
	}
}

func TestStringHashcode_DiffersBetweenInputs(t *testing.T) {
	if stringHashcode("a") == stringHashcode("b") {
		t.Fatal("stringHashcode returned identical hashes for different inputs")
	}
}

func TestSetDefaultRegionForImport(t *testing.T) {
	cases := []struct {
		name       string
		id         string
		wantRegion string
		wantId     string
	}{
		{"no prefix", "example.com", "us", "example.com"},
		{"us prefix", "us:example.com", "us", "example.com"},
		{"eu prefix", "eu:example.com", "eu", "example.com"},
		{"empty region falls back", ":example.com", "us", ":example.com"},
		{"empty id falls back", "us:", "us", "us:"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"region": {Type: schema.TypeString, Optional: true},
			}, map[string]interface{}{})
			d.SetId(tc.id)

			setDefaultRegionForImport(d)

			if got := d.Get("region").(string); got != tc.wantRegion {
				t.Errorf("region = %q, want %q", got, tc.wantRegion)
			}
			if got := d.Id(); got != tc.wantId {
				t.Errorf("id = %q, want %q", got, tc.wantId)
			}
		})
	}
}

func TestGenerateId(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceMailgunWebhook().Schema, map[string]interface{}{
		"region": "eu",
		"domain": "example.com",
		"kind":   "delivered",
	})

	got := generateId(d)
	want := "eu:example.com:delivered"
	if got != want {
		t.Errorf("generateId = %q, want %q", got, want)
	}
}

func TestWebhookKindValidator(t *testing.T) {
	validator := resourceMailgunWebhook().Schema["kind"].ValidateFunc

	allowed := []string{"accepted", "clicked", "complained", "delivered", "opened", "permanent_fail", "temporary_fail", "unsubscribed"}
	for _, kind := range allowed {
		_, errs := validator(kind, "kind")
		if len(errs) != 0 {
			t.Errorf("kind %q should be valid, got errors: %v", kind, errs)
		}
	}

	_, errs := validator("not-a-kind", "kind")
	if len(errs) == 0 {
		t.Errorf("invalid kind should report an error")
	}
}

func TestIsNotFound(t *testing.T) {
	if isNotFound(nil) {
		t.Errorf("nil error should not be treated as not-found")
	}
	if isNotFound(errors.New("boom")) {
		t.Errorf("plain error should not be treated as not-found")
	}

	notFound := &mailgun.UnexpectedResponseError{
		Expected: []int{200},
		Actual:   http.StatusNotFound,
	}
	if !isNotFound(notFound) {
		t.Errorf("UnexpectedResponseError with 404 should be detected")
	}

	wrapped := errors.New("wrap: " + notFound.Error())
	if isNotFound(wrapped) {
		t.Errorf("string-wrapped 404 (no errors.As link) should not match")
	}

	errorfWrapped := fmt.Errorf("retrieving foo: %w", notFound)
	if !isNotFound(errorfWrapped) {
		t.Errorf("fmt.Errorf %%w-wrapped 404 should be detected via errors.As")
	}

	doubleWrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", notFound))
	if !isNotFound(doubleWrapped) {
		t.Errorf("doubly %%w-wrapped 404 should be detected via errors.As")
	}

	other := &mailgun.UnexpectedResponseError{
		Expected: []int{200},
		Actual:   http.StatusInternalServerError,
	}
	if isNotFound(other) {
		t.Errorf("non-404 status should not match")
	}

	if isNotFound(fmt.Errorf("wrap: %w", other)) {
		t.Errorf("wrapped non-404 status should not match")
	}
}

func TestResourceMailgunWebhookImport(t *testing.T) {
	cases := []struct {
		name       string
		id         string
		wantRegion string
		wantDomain string
		wantKind   string
		wantErr    bool
	}{
		{"three parts eu", "eu:example.com:delivered", "eu", "example.com", "delivered", false},
		{"three parts us", "us:example.com:opened", "us", "example.com", "opened", false},
		{"two parts defaults to us", "example.com:clicked", "us", "example.com", "clicked", false},
		{"single part is invalid", "example.com", "", "", "", true},
		{"empty string is invalid", "", "", "", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceMailgunWebhook().Schema, map[string]interface{}{})
			d.SetId(tc.id)

			got, err := resourceMailgunWebhookImport(context.Background(), d, nil)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for id %q, got nil", tc.id)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != 1 || got[0] != d {
				t.Fatalf("expected single ResourceData echoed back, got %#v", got)
			}
			if v := d.Get("region").(string); v != tc.wantRegion {
				t.Errorf("region = %q, want %q", v, tc.wantRegion)
			}
			if v := d.Get("domain").(string); v != tc.wantDomain {
				t.Errorf("domain = %q, want %q", v, tc.wantDomain)
			}
			if v := d.Get("kind").(string); v != tc.wantKind {
				t.Errorf("kind = %q, want %q", v, tc.wantKind)
			}
		})
	}
}

// Domain CustomizeDiff has been replaced by ModifyPlan in the framework
// resource (internal/framework/domain_resource.go). Equivalent coverage lives
// in framework acceptance tests.
