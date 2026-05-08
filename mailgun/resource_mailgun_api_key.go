package mailgun

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

func resourceMailgunApiKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMailgunApiKeyCreate,
		DeleteContext: resourceMailgunApiKeyDelete,
		ReadContext:   resourceMailgunApiKeyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"kind": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "user",
			},
			"region": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "us",
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"requestor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"expires_at": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"is_disabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"disabled_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMailgunApiKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	opts := mailgun.CreateAPIKeyOptions{}

	role := d.Get("role").(string)

	opts.Description = d.Get("description").(string)
	opts.DomainName = d.Get("domain_name").(string)
	opts.Email = d.Get("email").(string)
	opts.Expiration = uint64(d.Get("expires_at").(int))
	opts.Kind = d.Get("kind").(string)
	opts.UserID = d.Get("user_id").(string)
	opts.UserName = d.Get("user_name").(string)

	apiKey, err := client.CreateAPIKey(ctx, role, &opts)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(apiKey.ID)
	_ = d.Set("requestor", apiKey.Requestor)
	_ = d.Set("secret", apiKey.Secret)

	log.Printf("[INFO] API key ID: %s", d.Id())

	return nil
}

func resourceMailgunApiKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	log.Printf("[INFO] Deleting API key: %s", d.Id())

	// Destroy the API key
	err := client.DeleteAPIKey(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Error deleting API key: %s", err)
	}

	return nil
}

func resourceMailgunApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	err := resourceMailgunApiKeyRetrieve(ctx, d.Id(), client, d)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMailgunApiKeyRetrieve(ctx context.Context, id string, client *mailgun.Client, d *schema.ResourceData) error {
	resp, err := client.ListAPIKeys(ctx, nil)

	if err != nil {
		return fmt.Errorf("Error retrieving API key list: %s", err)
	}

	var apiKey mtypes.APIKey

	for i, key := range resp {
		if resp[i].ID == id {
			apiKey = key
		}
	}

	if apiKey.ID == "" {
		log.Printf("[DEBUG] API key not found with ID: %s", d.Id())
		d.SetId("")
		return nil
	}

	applyAPIKey(d, apiKey)
	return nil
}

// applyAPIKey copies fields from apiKey into d. The secret is preserved when
// the API returns an empty value because Mailgun only returns it on creation
// (regression #73).
func applyAPIKey(d *schema.ResourceData, apiKey mtypes.APIKey) {
	_ = d.Set("description", apiKey.Description)
	_ = d.Set("kind", apiKey.Kind)
	_ = d.Set("role", apiKey.Role)
	_ = d.Set("domain_name", apiKey.DomainName)
	_ = d.Set("user_name", apiKey.UserName)
	_ = d.Set("requestor", apiKey.Requestor)
	_ = d.Set("is_disabled", apiKey.IsDisabled)
	_ = d.Set("disabled_reason", apiKey.DisabledReason)

	if apiKey.Secret != "" {
		_ = d.Set("secret", apiKey.Secret)
	}
}
