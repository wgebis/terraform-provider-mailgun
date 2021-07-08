package mailgun

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMailgunWebhook() *schema.Resource {
	log.Printf("[DEBUG] resourceMailgunWebhook()")

	return &schema.Resource{
		CreateContext: resourceMailgunWebhookCreate,
		ReadContext:   resourceMailgunWebhookRead,
		UpdateContext: resourceMailgunWebhookUpdate,
		DeleteContext: resourceMailgunWebhookDelete,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "us",
			},

			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"kind": {
				Type:     schema.TypeString,
				ForceNew: false,
				Required: true,
				ValidateDiagFunc: func(val interface{}, _ cty.Path) diag.Diagnostics {
					v := val.(string)
					alllowedKinds := []string{"clicked", "complained", "delivered", "opened", "permanent_fail", "temporary_fail", "unsubscribed"}
					matched := false
					for _, kind := range alllowedKinds {
						if kind == v {
							matched = true
						}
					}
					if !matched {
						return diag.Errorf("kind must be %s", alllowedKinds)
					}
					return nil
				},
			},

			"urls": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceMailgunWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClientForDomain(d.Get("region").(string), d.Get("domain").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	kind := d.Get("kind").(string)
	urls := d.Get("urls").(*schema.Set)

	stringUrls := []string{}
	for _, url := range urls.List() {
		stringUrls = append(stringUrls, url.(string))
	}

	err := client.CreateWebhook(ctx, kind, stringUrls)
	if err != nil {
		return diag.FromErr(err)
	}

	id := generateId(d)
	d.SetId(id)

	log.Printf("[INFO] Create webhook ID: %s", d.Id())

	return resourceMailgunWebhookRead(ctx, d, meta)
}

func resourceMailgunWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClientForDomain(d.Get("region").(string), d.Get("domain").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	kind := d.Get("kind").(string)
	urls := d.Get("urls").(*schema.Set)

	stringUrls := []string{}
	for _, url := range urls.List() {
		stringUrls = append(stringUrls, url.(string))
	}

	err := client.UpdateWebhook(ctx, kind, stringUrls)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Update webhook ID: %s", d.Id())

	return resourceMailgunWebhookRead(ctx, d, meta)
}

func resourceMailgunWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClientForDomain(d.Get("region").(string), d.Get("domain").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	kind := d.Get("kind").(string)

	err := client.DeleteWebhook(ctx, kind)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Delete webhook ID: %s", d.Id())

	return nil
}

func resourceMailgunWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClientForDomain(d.Get("region").(string), d.Get("domain").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	kind := d.Get("kind").(string)
	urls, err := client.GetWebhook(ctx, kind)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("kind", kind)
	d.Set("urls", urls)

	return nil
}

func generateId(d *schema.ResourceData) string {
	region := d.Get("region").(string)
	domain := d.Get("domain").(string)
	kind := d.Get("kind").(string)
	return fmt.Sprintf("%s:%s:%s", region, domain, kind)
}
