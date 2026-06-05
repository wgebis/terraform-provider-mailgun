package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// domainResourceModel mirrors the mailgun_domain resource state. Field tags
// must match the schema attribute names.
type domainResourceModel struct {
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
	ReceivingRecordsSet        types.Set    `tfsdk:"receiving_records_set"`
	SendingRecordsSet          types.Set    `tfsdk:"sending_records_set"`
}

// sendingRecordModel mirrors a sending_records_set element.
type sendingRecordModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	RecordType types.String `tfsdk:"record_type"`
	Valid      types.String `tfsdk:"valid"`
	Value      types.String `tfsdk:"value"`
}

// receivingRecordModel mirrors a receiving_records_set element.
type receivingRecordModel struct {
	ID         types.String `tfsdk:"id"`
	Priority   types.String `tfsdk:"priority"`
	RecordType types.String `tfsdk:"record_type"`
	Valid      types.String `tfsdk:"valid"`
	Value      types.String `tfsdk:"value"`
}

// sendingRecordObjectType describes the cty type of a sending_records_set
// element. It is used to construct types.Set values in the resource code.
func sendingRecordObjectType() attr.Type {
	return types.ObjectType{AttrTypes: sendingRecordAttrTypes()}
}

func sendingRecordAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"record_type": types.StringType,
		"valid":       types.StringType,
		"value":       types.StringType,
	}
}

// receivingRecordObjectType describes the cty type of a receiving_records_set
// element.
func receivingRecordObjectType() attr.Type {
	return types.ObjectType{AttrTypes: receivingRecordAttrTypes()}
}

func receivingRecordAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":          types.StringType,
		"priority":    types.StringType,
		"record_type": types.StringType,
		"valid":       types.StringType,
		"value":       types.StringType,
	}
}
