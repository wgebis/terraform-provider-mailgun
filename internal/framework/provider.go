// Package framework hosts the terraform-plugin-framework implementation of the
// Mailgun provider. It is muxed together with the legacy SDKv2 provider in
// main.go so resources can be migrated incrementally without breaking state
// compatibility.
package framework

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/wgebis/terraform-provider-mailgun/mailgun"
)

// Ensure the implementation satisfies the expected interfaces.
var _ provider.Provider = (*mailgunProvider)(nil)

// New returns a constructor for the framework provider. The constructor form
// is required by providerserver.NewProtocol6.
func New() func() provider.Provider {
	return func() provider.Provider {
		return &mailgunProvider{}
	}
}

type mailgunProvider struct{}

// providerModel mirrors the SDKv2 provider schema. The two schemas must stay
// in sync for the mux server to merge them without conflicts.
type providerModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func (p *mailgunProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mailgun"
}

func (p *mailgunProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *mailgunProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := data.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("MAILGUN_API_KEY")
	}

	cfg := &mailgun.Config{APIKey: apiKey}
	resp.DataSourceData = cfg
	resp.ResourceData = cfg
}

func (p *mailgunProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainResource,
	}
}

func (p *mailgunProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
	}
}
