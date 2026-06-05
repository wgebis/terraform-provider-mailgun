package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

var (
	_ datasource.DataSource              = (*domainDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*domainDataSource)(nil)
)

// NewDomainDataSource is the constructor registered with the framework
// provider for data "mailgun_domain".
func NewDomainDataSource() datasource.DataSource {
	return &domainDataSource{}
}

type domainDataSource struct {
	cfg *mailgunpkg.Config
}

func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *domainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Attributes: map[string]dsschema.Attribute{
			"id":                            dsschema.StringAttribute{Computed: true},
			"name":                          dsschema.StringAttribute{Required: true},
			"region":                        dsschema.StringAttribute{Optional: true, Computed: true},
			"spam_action":                   dsschema.StringAttribute{Computed: true},
			"smtp_login":                    dsschema.StringAttribute{Computed: true},
			"smtp_password":                 dsschema.StringAttribute{Computed: true, Sensitive: true},
			"wildcard":                      dsschema.BoolAttribute{Computed: true},
			"dkim_selector":                 dsschema.StringAttribute{Computed: true},
			"force_dkim_authority":          dsschema.BoolAttribute{Computed: true},
			"open_tracking":                 dsschema.BoolAttribute{Computed: true},
			"click_tracking":                dsschema.BoolAttribute{Computed: true},
			"web_scheme":                    dsschema.StringAttribute{Computed: true},
			"dkim_key_size":                 dsschema.Int64Attribute{Computed: true},
			"use_automatic_sender_security": dsschema.BoolAttribute{Computed: true},
			"sending_records_set":           dsSendingRecordsSetAttribute(),
			"receiving_records_set":         dsReceivingRecordsSetAttribute(),
		},
	}
}

func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(*mailgunpkg.Config)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data",
			fmt.Sprintf("expected *mailgun.Config, got %T", req.ProviderData))
		return
	}
	d.cfg = cfg
}

func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data domainResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := data.Region.ValueString()
	if region == "" {
		region = "us"
	}
	client, err := d.cfg.GetClient(region)
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	name := data.Name.ValueString()
	data.ID = types.StringValue(name)
	data.Region = types.StringValue(region)

	diags, notFound := refreshDomain(ctx, client, name, &data)
	if notFound {
		resp.Diagnostics.AddError("Domain not found",
			fmt.Sprintf("Mailgun domain %q does not exist", name))
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func dsSendingRecordsSetAttribute() dsschema.SetNestedAttribute {
	return dsschema.SetNestedAttribute{
		Computed: true,
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: map[string]dsschema.Attribute{
				"id":          dsschema.StringAttribute{Computed: true},
				"name":        dsschema.StringAttribute{Computed: true},
				"record_type": dsschema.StringAttribute{Computed: true},
				"valid":       dsschema.StringAttribute{Computed: true},
				"value":       dsschema.StringAttribute{Computed: true},
			},
		},
	}
}

func dsReceivingRecordsSetAttribute() dsschema.SetNestedAttribute {
	return dsschema.SetNestedAttribute{
		Computed: true,
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: map[string]dsschema.Attribute{
				"id":          dsschema.StringAttribute{Computed: true},
				"priority":    dsschema.StringAttribute{Computed: true},
				"record_type": dsschema.StringAttribute{Computed: true},
				"valid":       dsschema.StringAttribute{Computed: true},
				"value":       dsschema.StringAttribute{Computed: true},
			},
		},
	}
}
