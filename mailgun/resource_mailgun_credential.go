package mailgun

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/mailgun/mailgun-go/v3"

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
				ForceNew: true,
			},

			"password": {
				Type:     schema.TypeString,
				ForceNew: false,
				Required: true,
			},

			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
	parts := strings.SplitN(d.Id(), ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		d.Set("region", "us")
	} else {
		d.Set("region", parts[0])
		d.SetId(parts[1])
	}

	log.Printf("[DEBUG] Import credential for region '%s' and email '%s'", d.Get("region"), d.Id())

	return []*schema.ResourceData{d}, nil
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

	log.Printf("[INFO] Create credential ID: %s", d.Id())

	return nil
}

func resourceMailgunCredentialUpdate(d *schema.ResourceData, meta interface{}) error {
	client, errc := meta.(*Config).GetClientForDomain(d.Get("region").(string), d.Get("domain").(string))
	if errc != nil {
		return errc
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

	log.Printf("[DEBUG] Credential update configuration: %#v", cred)

	err := client.ChangeCredentialPassword(context.Background(), email, password)

	if err != nil {
		return err
	}

	d.SetId(email)

	log.Printf("[INFO] Update credential ID: %s", d.Id())

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
	parts := strings.SplitN(d.Id(), "@", 2)

	if len(parts) != 2 {
		return fmt.Errorf("The ID of credential '%s' don't contains domain!", d.Id())
	}

	domain := parts[1]

	client, errc := meta.(*Config).GetClientForDomain(d.Get("region").(string), domain)
	if errc != nil {
		return errc
	}

	log.Printf("[DEBUG] Read credential for region '%s' and email '%s'", d.Get("region"), d.Id())

	itCredentials := client.ListCredentials(nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var page []mailgun.Credential

	for itCredentials.Next(ctx, &page) {
		log.Printf("[DEBUG] Read credential get new page")

		for _, c := range page {
			if c.Login == d.Id() {
				d.Set("email", c.Login)
				d.Set("domain", domain)
				d.Set("password", c.Password)

				return nil
			}
		}
	}

	if err := itCredentials.Err(); err != nil {
		return err
	}

	return fmt.Errorf("The credential '%s' not found!", d.Id())
}
