package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// domainResourceModelV0 is the prior shape of the resource state, as written
// by the legacy SDKv2 implementation. It carries the deprecated
// sending_records / receiving_records list attributes that are dropped in
// schema version 1.
type domainResourceModelV0 struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Region                     types.String `tfsdk:"region"`
	SpamAction                 types.String `tfsdk:"spam_action"`
	SmtpLogin                  types.String `tfsdk:"smtp_login"`
	SmtpPassword               types.String `tfsdk:"smtp_password"`
	Wildcard                   types.Bool   `tfsdk:"wildcard"`
	DkimSelector               types.String `tfsdk:"dkim_selector"`
	ForceDkimAuthority         types.Bool   `tfsdk:"force_dkim_authority"`
	OpenTracking               types.Bool   `tfsdk:"open_tracking"`
	ClickTracking              types.Bool   `tfsdk:"click_tracking"`
	WebScheme                  types.String `tfsdk:"web_scheme"`
	DkimKeySize                types.Int64  `tfsdk:"dkim_key_size"`
	UseAutomaticSenderSecurity types.Bool   `tfsdk:"use_automatic_sender_security"`
	ReceivingRecords           types.List   `tfsdk:"receiving_records"`
	ReceivingRecordsSet        types.Set    `tfsdk:"receiving_records_set"`
	SendingRecords             types.List   `tfsdk:"sending_records"`
	SendingRecordsSet          types.Set    `tfsdk:"sending_records_set"`
}

// toV1 strips the deprecated list attributes; the remaining values are copied
// verbatim into the new model.
func (v domainResourceModelV0) toV1() domainResourceModel {
	return domainResourceModel{
		ID:                         v.ID,
		Name:                       v.Name,
		Region:                     v.Region,
		SpamAction:                 v.SpamAction,
		SmtpLogin:                  v.SmtpLogin,
		SmtpPassword:               v.SmtpPassword,
		Wildcard:                   v.Wildcard,
		DkimSelector:               v.DkimSelector,
		ForceDkimAuthority:         v.ForceDkimAuthority,
		OpenTracking:               v.OpenTracking,
		ClickTracking:              v.ClickTracking,
		WebScheme:                  v.WebScheme,
		DkimKeySize:                v.DkimKeySize,
		UseAutomaticSenderSecurity: v.UseAutomaticSenderSecurity,
		ReceivingRecordsSet:        v.ReceivingRecordsSet,
		SendingRecordsSet:          v.SendingRecordsSet,
	}
}

// domainSchemaV0 returns a minimal prior schema describing the v0 state
// shape. Only attribute types are required for the upgrader; defaults and
// plan modifiers are not consulted during state upgrade.
func domainSchemaV0() *schema.Schema {
	return &schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                            schema.StringAttribute{Computed: true},
			"name":                          schema.StringAttribute{Required: true},
			"region":                        schema.StringAttribute{Optional: true, Computed: true},
			"spam_action":                   schema.StringAttribute{Optional: true, Computed: true},
			"smtp_login":                    schema.StringAttribute{Computed: true},
			"smtp_password":                 schema.StringAttribute{Optional: true, Sensitive: true},
			"wildcard":                      schema.BoolAttribute{Optional: true, Computed: true},
			"dkim_selector":                 schema.StringAttribute{Optional: true},
			"force_dkim_authority":          schema.BoolAttribute{Optional: true},
			"open_tracking":                 schema.BoolAttribute{Optional: true, Computed: true},
			"click_tracking":                schema.BoolAttribute{Optional: true, Computed: true},
			"web_scheme":                    schema.StringAttribute{Optional: true, Computed: true},
			"dkim_key_size":                 schema.Int64Attribute{Optional: true},
			"use_automatic_sender_security": schema.BoolAttribute{Optional: true, Computed: true},

			"receiving_records":     deprecatedListAttr([]string{"id", "priority", "record_type", "valid", "value"}),
			"receiving_records_set": receivingRecordsSetAttribute(),
			"sending_records":       deprecatedListAttr([]string{"id", "name", "record_type", "valid", "value"}),
			"sending_records_set":   sendingRecordsSetAttribute(),
		},
	}
}

func deprecatedListAttr(fields []string) schema.ListNestedAttribute {
	attrs := map[string]schema.Attribute{}
	for _, f := range fields {
		attrs[f] = schema.StringAttribute{Computed: true}
	}
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: attrs,
		},
	}
}
