package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/mailgun/mailgun-go/v5"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

// refreshDomain fetches the latest domain + tracking from Mailgun and applies
// the result to the supplied model. The returned bool indicates whether the
// API responded with 404 (domain has been deleted out-of-band).
func refreshDomain(ctx context.Context, client *mailgun.Client, id string, m *domainResourceModel) (diag.Diagnostics, bool) {
	var diags diag.Diagnostics

	resp, err := client.GetDomain(ctx, id, nil)
	if err != nil {
		if mailgunpkg.IsNotFound(err) {
			return diags, true
		}
		diags.AddError("Failed to read domain", fmt.Sprintf("error retrieving domain %q: %s", id, err))
		return diags, false
	}

	tracking, err := client.GetDomainTracking(ctx, id)
	if err != nil {
		diags.AddError("Failed to read domain tracking", fmt.Sprintf("error retrieving tracking for %q: %s", id, err))
		return diags, false
	}

	diags.Append(applyDomainResponse(ctx, m, &resp, &tracking)...)
	return diags, false
}
