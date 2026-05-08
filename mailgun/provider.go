package mailgun

import (
	"context"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},

		// All resources and data sources are served by the framework provider
		// via tf6muxserver. The SDKv2 provider remains only to satisfy the
		// mux until the SDKv2 dependency is removed entirely.
		DataSourcesMap: map[string]*schema.Resource{},
		ResourcesMap:   map[string]*schema.Resource{},
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(d)
	}

	return p
}

func providerConfigure(d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	if apiKey == "" {
		apiKey = os.Getenv("MAILGUN_API_KEY")
	}
	if apiKey == "" {
		return nil, diag.Errorf("api_key is required: set the api_key provider argument or the MAILGUN_API_KEY environment variable")
	}

	config := Config{APIKey: apiKey}

	log.Println("[INFO] Initializing Mailgun client")
	return config.Client()
}
