package mailgun

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/mailgun/mailgun-go/v5"
)

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
