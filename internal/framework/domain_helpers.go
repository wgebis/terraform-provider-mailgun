package framework

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// applyDomainResponse fills the model fields with values from a Mailgun
// GetDomainResponse. The smtp_password field is intentionally not touched
// because the Mailgun API never returns it.
func applyDomainResponse(ctx context.Context, m *domainResourceModel, resp *mtypes.GetDomainResponse, tracking *mtypes.DomainTracking) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Name = types.StringValue(resp.Domain.Name)
	m.SmtpLogin = types.StringValue(resp.Domain.SMTPLogin)
	m.Wildcard = types.BoolValue(resp.Domain.Wildcard)
	m.SpamAction = types.StringValue(string(resp.Domain.SpamAction))
	m.WebScheme = types.StringValue(resp.Domain.WebScheme)
	m.UseAutomaticSenderSecurity = types.BoolValue(resp.Domain.UseAutomaticSenderSecurity)

	sending := make([]sendingRecordModel, len(resp.SendingDNSRecords))
	for i, r := range resp.SendingDNSRecords {
		id := r.Name
		if strings.Contains(r.Name, "._domainkey.") && !resp.Domain.UseAutomaticSenderSecurity {
			id = "_domainkey." + resp.Domain.Name
		}
		sending[i] = sendingRecordModel{
			ID:         types.StringValue(id),
			Name:       types.StringValue(r.Name),
			RecordType: types.StringValue(r.RecordType),
			Valid:      types.StringValue(r.Valid),
			Value:      types.StringValue(r.Value),
		}
	}
	sendingSet, d := types.SetValueFrom(ctx, sendingRecordObjectType(), sending)
	diags.Append(d...)
	if !diags.HasError() {
		m.SendingRecordsSet = sendingSet
	}

	receiving := make([]receivingRecordModel, len(resp.ReceivingDNSRecords))
	for i, r := range resp.ReceivingDNSRecords {
		receiving[i] = receivingRecordModel{
			ID:         types.StringValue(r.Value),
			Priority:   types.StringValue(r.Priority),
			RecordType: types.StringValue(r.RecordType),
			Valid:      types.StringValue(r.Valid),
			Value:      types.StringValue(r.Value),
		}
	}
	receivingSet, d := types.SetValueFrom(ctx, receivingRecordObjectType(), receiving)
	diags.Append(d...)
	if !diags.HasError() {
		m.ReceivingRecordsSet = receivingSet
	}

	if tracking != nil {
		m.OpenTracking = types.BoolValue(tracking.Open.Active)
		m.ClickTracking = types.BoolValue(tracking.Click.Active)
	}

	return diags
}
