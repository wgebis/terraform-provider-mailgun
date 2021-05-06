package mailgun

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mailgun/mailgun-go/v3"
)

func resourceMailgunCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMailgunCredentialCreate,
		ReadContext:   resourceMailgunCredentialRead,
		DeleteContext: resourceMailgunCredentialDelete,
		UpdateContext: resourceMailgunCredentialUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMailgunCredentialImport,
		},
		Schema: map[string]*schema.Schema{
			"login": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceMailgunCredentialImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	parts := strings.SplitN(d.Id(), ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("Failed to import domain credentials, invalid format")
	} else {
		d.Set("domain", parts[0])
		d.Set("login", parts[1])
	}

	return []*schema.ResourceData{d}, nil
}

func resourceMailgunCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domain := d.Get("domain").(string)
	password := d.Get("password").(string)
	login := d.Get("login").(string)
	client, errc := meta.(*Config).GetClientForDomain("", domain)
	if errc != nil {
		return diag.FromErr(errc)
	}

	errc = client.ChangeCredentialPassword(ctx, login, password)
	if errc != nil {
		return diag.FromErr(errc)
	}

	return nil
}

func resourceMailgunCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domain := d.Get("domain").(string)
	client, errc := meta.(*Config).GetClientForDomain("", domain)
	if errc != nil {
		return diag.FromErr(errc)
	}
	login := d.Get("login").(string)
	password := d.Get("password").(string)
	log.Printf("[DEBUG] create credential with login: %s", login)
	err := client.CreateCredential(ctx, login, password)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s:%s", domain, login))

	log.Printf("[INFO] credential ID: %s", d.Id())

	return nil
}

func resourceMailgunCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClientForDomain("", d.Get("domain").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	log.Printf("[INFO] Deleting credential: %s", d.Id())

	err := client.DeleteCredential(ctx, d.Get("login").(string))
	if err != nil {
		return diag.Errorf("Error deleting credential: %s", err)
	}

	return nil
}

func resourceMailgunCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domain := d.Get("domain").(string)
	login := d.Get("login").(string)
	client, errc := meta.(*Config).GetClientForDomain("", domain)
	if errc != nil {
		return diag.FromErr(errc)
	}
	it := client.ListCredentials(nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var page []mailgun.Credential
	for it.Next(ctx, &page) {
		for _, item := range page {
			if item.Login == login {
				return nil
			}
		}
	}

	if it.Err() != nil {
		return diag.FromErr(it.Err())
	}

	return nil
}
