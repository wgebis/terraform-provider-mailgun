package mailgun

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("MAILGUN_API_KEY", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"mailgun_domain": dataSourceMailgunDomain(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"mailgun_domain": resourceMailgunDomain(),
			"mailgun_route":  resourceMailgunRoute(),
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		return providerConfigure(d)
	}

	return p
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		APIKey: d.Get("api_key").(string),
	}

	log.Println("[INFO] Initializing Mailgun client")
	return config.Client()
}
