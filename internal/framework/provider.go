// Package framework hosts the terraform-plugin-framework implementation of the
// Mailgun provider. It serves every resource and data source over protocol v6
// and is the sole runtime for the provider.
package framework

import (
	"context"
	"os"
	"strconv"

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
	APIKey            types.String  `tfsdk:"api_key"`
	RequestsPerSecond types.Float64 `tfsdk:"requests_per_second"`
	MaxRetries        types.Int64   `tfsdk:"max_retries"`
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
			"requests_per_second": schema.Float64Attribute{
				Optional: true,
				Description: "Maximum Mailgun API requests per second for this provider instance. " +
					"0 or unset disables pacing. Can also be set with MAILGUN_REQUESTS_PER_SECOND.",
			},
			"max_retries": schema.Int64Attribute{
				Optional: true,
				Description: "Number of times to retry a request that receives HTTP 429. " +
					"0 or unset disables retries. Can also be set with MAILGUN_MAX_RETRIES.",
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

	rps := data.RequestsPerSecond.ValueFloat64()
	if data.RequestsPerSecond.IsNull() {
		if v := os.Getenv("MAILGUN_REQUESTS_PER_SECOND"); v != "" {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				resp.Diagnostics.AddError("Invalid MAILGUN_REQUESTS_PER_SECOND",
					"MAILGUN_REQUESTS_PER_SECOND must be a number: "+err.Error())
				return
			}
			rps = f
		}
	}
	maxRetries := data.MaxRetries.ValueInt64()
	if data.MaxRetries.IsNull() {
		if v := os.Getenv("MAILGUN_MAX_RETRIES"); v != "" {
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				resp.Diagnostics.AddError("Invalid MAILGUN_MAX_RETRIES",
					"MAILGUN_MAX_RETRIES must be an integer: "+err.Error())
				return
			}
			maxRetries = n
		}
	}
	// !(rps >= 0) also rejects NaN, which ParseFloat accepts.
	if !(rps >= 0) || maxRetries < 0 {
		resp.Diagnostics.AddError("Invalid provider configuration",
			"requests_per_second and max_retries must not be negative")
		return
	}

	cfg := &mailgun.Config{
		APIKey:            apiKey,
		RequestsPerSecond: rps,
		MaxRetries:        int(maxRetries),
	}
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
