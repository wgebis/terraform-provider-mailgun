package mailgun

import (
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

// Webhook import + kind validator coverage lives in the framework resource
// (internal/framework/webhook_resource.go) and its acceptance tests.
//
// Domain CustomizeDiff has been replaced by ModifyPlan in the framework
// resource (internal/framework/domain_resource.go). Equivalent coverage lives
// in framework acceptance tests.
