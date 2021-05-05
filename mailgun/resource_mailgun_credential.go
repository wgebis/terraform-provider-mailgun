package mailgun

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type mailgunCredential struct {
	Email    string
	Region   string
	Domain   string
	Password string
}

func resourceMailgunCredential() *schema.Resource {
	log.Printf("[DEBUG] resourceMailgunCredential()")

	return &schema.Resource{
		CreateContext: resourceMailgunCredentialCreate,
		Read:          resourceMailgunCredentialRead,
		Update:        resourceMailgunCredentialUpdate,
		Delete:        resourceMailgunCredentialDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMailgunCredentialImport,
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"password": {
				Type:     schema.TypeString,
				ForceNew: false,
				Optional: true,
			},

			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"region": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "us",
			},
		},
	}
}

func resourceMailgunCredentialImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// see route

	log.Printf("[DEBUG] resourceMailgunCredentialImport()")

	return nil, nil
}

func resourceMailgunCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClientForDomain(d.Get("region").(string), d.Get("domain").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	email := d.Get("email").(string)
	password := d.Get("password").(string)
	domain := d.Get("domain").(string)
	region := d.Get("region").(string)

	cred := mailgunCredential{
		Email:    email,
		Password: "****",
		Domain:   domain,
		Region:   region,
	}

	log.Printf("[DEBUG] Credential create configuration: %#v", cred)

	err := client.CreateCredential(context.Background(), email, password)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(email)

	log.Printf("[INFO] Credential ID: %s", d.Id())

	return nil
}

func resourceMailgunCredentialUpdate(d *schema.ResourceData, meta interface{}) error {
	// client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	// if errc != nil {
	// 	return errc
	// }
	// see route

	log.Printf("[DEBUG] resourceMailgunCredentialUpdate()")

	return nil
}

func resourceMailgunCredentialDelete(d *schema.ResourceData, meta interface{}) error {
	client, errc := meta.(*Config).GetClientForDomain(d.Get("region").(string), d.Get("domain").(string))
	if errc != nil {
		return errc
	}

	email := d.Get("email").(string)
	err := client.DeleteCredential(context.Background(), email)

	if err != nil {
		return fmt.Errorf("Error deleting route: %s", err)
	}

	return nil
}

func resourceMailgunCredentialRead(d *schema.ResourceData, meta interface{}) error {
	// client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	// if errc != nil {
	// 	return errc
	// }
	// see route

	log.Printf("[DEBUG] resourceMailgunCredentialRead()")

	return nil
}
