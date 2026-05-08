package mailgun

import (
	"context"
	"fmt"
	"log"
	"strings"

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
		Importer: &schema.ResourceImporter{
			StateContext: resourceMailgunWebhookImport,
		},

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
				ForceNew: true,
				Required: true,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					allowedKinds := []string{"accepted", "clicked", "complained", "delivered", "opened", "permanent_fail", "temporary_fail", "unsubscribed"}
					matched := false
					for _, kind := range allowedKinds {
						if kind == v {
							matched = true
						}
					}
					if !matched {
						errs = append(errs, fmt.Errorf("kind must be %s", allowedKinds))
					}
					return
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

func resourceMailgunWebhookImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[DEBUG] Import ID: %s", d.Id())
	parts := strings.SplitN(d.Id(), ":", 3)
	log.Printf("[DEBUG] Split parts: %v", parts)

	var region, domain, kind string

	switch len(parts) {
	case 2:
		region = "us"
		domain = parts[0]
		kind = parts[1]
	case 3:
		region = parts[0]
		domain = parts[1]
		kind = parts[2]
	default:
		return nil, fmt.Errorf("invalid import ID format. Expected 'region:domain:kind' or 'domain:kind'")
	}

	log.Printf("[DEBUG] Setting region=%s, domain=%s, kind=%s", region, domain, kind)
	_ = d.Set("region", region)
	_ = d.Set("domain", domain)
	_ = d.Set("kind", kind)

	log.Printf("[DEBUG] After setting - region: %s, domain: %s, kind: %s", d.Get("region"), d.Get("domain"), d.Get("kind"))
	return []*schema.ResourceData{d}, nil
}

func resourceMailgunWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	kind := d.Get("kind").(string)
	urls := d.Get("urls").(*schema.Set)

	stringUrls := []string{}
	for _, url := range urls.List() {
		stringUrls = append(stringUrls, url.(string))
	}

	err := client.CreateWebhook(ctx, d.Get("domain").(string), kind, stringUrls)
	if err != nil {
		return diag.FromErr(err)
	}

	id := generateId(d)
	d.SetId(id)

	log.Printf("[INFO] Create webhook ID: %s", d.Id())

	return resourceMailgunWebhookRead(ctx, d, meta)
}

func resourceMailgunWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	kind := d.Get("kind").(string)
	urls := d.Get("urls").(*schema.Set)

	stringUrls := []string{}
	for _, url := range urls.List() {
		stringUrls = append(stringUrls, url.(string))
	}

	err := client.UpdateWebhook(ctx, d.Get("domain").(string), kind, stringUrls)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Update webhook ID: %s", d.Id())

	return resourceMailgunWebhookRead(ctx, d, meta)
}

func resourceMailgunWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	kind := d.Get("kind").(string)

	err := client.DeleteWebhook(ctx, d.Get("domain").(string), kind)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Delete webhook ID: %s", d.Id())

	return nil
}

func resourceMailgunWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	kind := d.Get("kind").(string)
	urls, err := client.GetWebhook(ctx, d.Get("domain").(string), kind)
	if err != nil {
		if isNotFound(err) {
			log.Printf("[WARN] Mailgun webhook %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("kind", kind)
	_ = d.Set("urls", urls)

	return nil
}

func generateId(d *schema.ResourceData) string {
	region := d.Get("region").(string)
	domain := d.Get("domain").(string)
	kind := d.Get("kind").(string)
	return fmt.Sprintf("%s:%s:%s", region, domain, kind)
}
