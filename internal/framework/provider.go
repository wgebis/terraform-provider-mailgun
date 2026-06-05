// Package framework hosts the terraform-plugin-framework implementation of the
// Mailgun provider. It serves every resource and data source over protocol v6
// and is the sole runtime for the provider.
package framework

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

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

// NewProviderServer returns a protocol v6 provider server backed by the
// terraform-plugin-framework Mailgun provider. It is consumed by the binary
// entrypoint in main.go and by acceptance tests.
func NewProviderServer() (tfprotov6.ProviderServer, error) {
	return providerserver.NewProtocol6WithError(New()())()
}

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
		NewRouteResource,
		NewCredentialResource,
		NewWebhookResource,
		NewAPIKeyResource,
	}
}

func (p *mailgunProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
	}
}
