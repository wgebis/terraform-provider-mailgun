package mailgun

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("MAILGUN_API_KEY", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"mailgun_domain": dataSourceMailgunDomain(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"mailgun_domain":            resourceMailgunDomain(),
			"mailgun_route":             resourceMailgunRoute(),
			"mailgun_domain_credential": resourceMailgunCredential(),
			"mailgun_webhook":           resourceMailgunWebhook(),
		},
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(d)
	}

	return p
}

func providerConfigure(d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := Config{
		APIKey: d.Get("api_key").(string),
	}

	log.Println("[INFO] Initializing Mailgun client")
	return config.Client()
}
