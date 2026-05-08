package mailgun

import (
	"errors"
	"net/http"

	"github.com/mailgun/mailgun-go/v5"
)

// isNotFound reports whether err originated from a Mailgun API 404 response.
// Used by Read paths to clear the resource id so Terraform plans a recreate
// instead of failing on resources that were deleted out-of-band (issue #49).
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var ure *mailgun.UnexpectedResponseError
	if errors.As(err, &ure) {
		return ure.Actual == http.StatusNotFound
	}
	return mailgun.GetStatusFromErr(err) == http.StatusNotFound
}
