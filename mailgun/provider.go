package mailgun

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &mailgunProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &mailgunProvider{}
}

// mailgunProvider is the provider implementation.
type mailgunProvider struct{}

// mailgunProviderModel maps provider schema data to a Go type.
type mailgunProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

// Metadata returns the provider type name.
func (p *mailgunProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mailgun"
}

// Schema defines the provider-level schema for configuration data.
func (p *mailgunProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Mailgun.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "The Mailgun API key. May also be provided via MAILGUN_API_KEY environment variable.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a Mailgun API client for data sources and resources.
func (p *mailgunProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config mailgunProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unknown Mailgun API Key",
			"The provider cannot create the Mailgun API client as there is an unknown configuration value for the Mailgun API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MAILGUN_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	apiKey := os.Getenv("MAILGUN_API_KEY")

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Mailgun API Key Configuration",
			"While configuring the provider, the API key was not found in "+
				"the MAILGUN_API_KEY environment variable or provider "+
				"configuration block api_key attribute.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Mailgun client using the configuration values
	client := &Config{
		APIKey: apiKey,
	}

	// Make the Mailgun client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *mailgunProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *mailgunProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainResource,
		NewApiKeyResource,
		NewRouteResource,
		NewCredentialResource,
		NewWebhookResource,
	}
}
