package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

// webhookURLs flattens the urls Set into a []string for the Mailgun client.
func webhookURLs(ctx context.Context, m *webhookResourceModel) ([]string, diag.Diagnostics) {
	var urls []string
	d := m.URLs.ElementsAs(ctx, &urls, false)
	return urls, d
}

// webhookID composes the canonical "region:domain:kind" identifier used by
// the resource and its importer.
func webhookID(m *webhookResourceModel) string {
	return fmt.Sprintf("%s:%s:%s", m.Region.ValueString(), m.Domain.ValueString(), m.Kind.ValueString())
}

// refreshWebhook re-reads the webhook from Mailgun and updates the urls Set
// in the model. Used by Create/Update to align state with the API response.
func refreshWebhook(ctx context.Context, client *mailgun.Client, m *webhookResourceModel) diag.Diagnostics {
	urls, err := client.GetWebhook(ctx, m.Domain.ValueString(), m.Kind.ValueString())
	if err != nil {
		var diags diag.Diagnostics
		if mailgunpkg.IsNotFound(err) {
			diags.AddError("Webhook missing after write",
				fmt.Sprintf("webhook %s disappeared between write and read", webhookID(m)))
			return diags
		}
		diags.AddError("Failed to refresh webhook", err.Error())
		return diags
	}
	urlSet, d := types.SetValueFrom(ctx, types.StringType, urls)
	if d.HasError() {
		return d
	}
	m.URLs = urlSet
	return nil
}
