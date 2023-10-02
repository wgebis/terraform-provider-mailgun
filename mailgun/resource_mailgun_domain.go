package mailgun

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mailgun/mailgun-go/v4"
)

func resourceMailgunDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMailgunDomainCreate,
		ReadContext:   resourceMailgunDomainRead,
		UpdateContext: resourceMailgunDomainUpdate,
		DeleteContext: resourceMailgunDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMailgunDomainImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
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

			"spam_action": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "disabled",
			},

			"smtp_login": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"smtp_password": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Sensitive: true,
			},

			"wildcard": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
			},

			"dkim_selector": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"force_dkim_authority": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},

			"open_tracking": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
			},

			"receiving_records": &schema.Schema{
				Type:       schema.TypeList,
				Computed:   true,
				Deprecated: "Use `receiving_records_set` instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"priority": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"record_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"valid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"receiving_records_set": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Set:      domainRecordsSchemaSetFunc,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"priority": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"record_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"valid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"sending_records": &schema.Schema{
				Type:       schema.TypeList,
				Computed:   true,
				Deprecated: "Use `sending_records_set` instead.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"record_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"valid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"sending_records_set": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Set:      domainRecordsSchemaSetFunc,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"record_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"valid": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dkim_key_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
		},
		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				if diff.HasChange("name") {
					var sendingRecords []interface{}

					sendingRecords = append(sendingRecords, map[string]interface{}{"id": diff.Get("name").(string)})
					sendingRecords = append(sendingRecords, map[string]interface{}{"id": "_domainkey." + diff.Get("name").(string)})
					sendingRecords = append(sendingRecords, map[string]interface{}{"id": "email." + diff.Get("name").(string)})

					if err := diff.SetNew("sending_records_set", schema.NewSet(domainRecordsSchemaSetFunc, sendingRecords)); err != nil {
						return fmt.Errorf("error setting new sending_records_set diff: %w", err)
					}

					var receivingRecords []interface{}

					receivingRecords = append(receivingRecords, map[string]interface{}{"id": "mxa.mailgun.org"})
					receivingRecords = append(receivingRecords, map[string]interface{}{"id": "mxb.mailgun.org"})

					if err := diff.SetNew("receiving_records_set", schema.NewSet(domainRecordsSchemaSetFunc, receivingRecords)); err != nil {
						return fmt.Errorf("error setting new receiving_records_set diff: %w", err)
					}
				}

				return nil
			},
		),
	}
}

func domainRecordsSchemaSetFunc(v interface{}) int {
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m["id"].(string); ok {
		return stringHashcode(v)
	}

	return 0
}

func resourceMailgunDomainImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	setDefaultRegionForImport(d)

	return []*schema.ResourceData{d}, nil
}

func resourceMailgunDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClientForDomain(d.Get("region").(string), d.Get("name").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	var currentData schema.ResourceData
	var newPassword string = d.Get("smtp_password").(string)
	var smtpLogin string = d.Get("smtp_login").(string)
	var openTracking = d.Get("open_tracking").(bool)

	// Retrieve and update state of domain
	_, errc = resourceMailgunDomainRetrieve(d.Id(), client, &currentData)

	if errc != nil {
		return diag.FromErr(errc)
	}

	// Update default credential if changed
	if currentData.Get("smtp_password") != newPassword {
		errc = client.ChangeCredentialPassword(ctx, smtpLogin, newPassword)

		if errc != nil {
			return diag.FromErr(errc)
		}
	}

	if currentData.Get("open_tracking") != openTracking {
		var openTrackingValue = "no"
		if openTracking {
			openTrackingValue = "yes"
		}
		errc = client.UpdateOpenTracking(ctx, d.Get("name").(string), openTrackingValue)

		if errc != nil {
			return diag.FromErr(errc)
		}
	}

	return nil
}

func resourceMailgunDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	opts := mailgun.CreateDomainOptions{}

	name := d.Get("name").(string)

	opts.SpamAction = mailgun.SpamAction(d.Get("spam_action").(string))
	opts.Password = d.Get("smtp_password").(string)
	opts.Wildcard = d.Get("wildcard").(bool)
	opts.DKIMKeySize = d.Get("dkim_key_size").(int)
	opts.ForceDKIMAuthority = d.Get("force_dkim_authority").(bool)
	var dkimSelector = d.Get("dkim_selector").(string)
	var openTracking = d.Get("open_tracking").(bool)

	log.Printf("[DEBUG] Domain create configuration: %#v", opts)

	_, err := client.CreateDomain(context.Background(), name, &opts)

	if err != nil {
		return diag.FromErr(err)
	}

	if dkimSelector != "" {
		errc = client.UpdateDomainDkimSelector(ctx, d.Get("name").(string), dkimSelector)

		if errc != nil {
			return diag.FromErr(errc)
		}
	}
	if openTracking {
		errc = client.UpdateOpenTracking(ctx, d.Get("name").(string), "yes")

		if errc != nil {
			return diag.FromErr(errc)
		}
	}

	d.SetId(name)

	log.Printf("[INFO] Domain ID: %s", d.Id())

	// Retrieve and update state of domain
	_, err = resourceMailgunDomainRetrieve(d.Id(), client, d)

	if err != nil {
		return diag.FromErr(errc)
	}

	return nil
}

func resourceMailgunDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	log.Printf("[INFO] Deleting Domain: %s", d.Id())

	// Destroy the domain
	err := client.DeleteDomain(context.Background(), d.Id())
	if err != nil {
		return diag.Errorf("Error deleting domain: %s", err)
	}

	// Give the destroy a chance to take effect
	err = resource.RetryContext(ctx, 5*time.Minute, func() *resource.RetryError {
		_, err = client.GetDomain(ctx, d.Id())
		if err == nil {
			log.Printf("[INFO] Retrying until domain disappears...")
			return resource.RetryableError(
				fmt.Errorf("domain seems to still exist; will check again"))
		}
		log.Printf("[INFO] Got error looking for domain, seems gone: %s", err)
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMailgunDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	_, err := resourceMailgunDomainRetrieve(d.Id(), client, d)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMailgunDomainRetrieve(id string, client *mailgun.MailgunImpl, d *schema.ResourceData) (*mailgun.DomainResponse, error) {

	resp, err := client.GetDomain(context.Background(), id)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving domain: %s", err)
	}

	d.Set("name", resp.Domain.Name)
	d.Set("smtp_login", resp.Domain.SMTPLogin)
	d.Set("wildcard", resp.Domain.Wildcard)
	d.Set("spam_action", resp.Domain.SpamAction)

	receivingRecords := make([]map[string]interface{}, len(resp.ReceivingDNSRecords))
	for i, r := range resp.ReceivingDNSRecords {
		receivingRecords[i] = make(map[string]interface{})
		receivingRecords[i]["id"] = r.Value
		receivingRecords[i]["priority"] = r.Priority
		receivingRecords[i]["valid"] = r.Valid
		receivingRecords[i]["value"] = r.Value
		receivingRecords[i]["record_type"] = r.RecordType
	}
	d.Set("receiving_records", receivingRecords)
	d.Set("receiving_records_set", receivingRecords)

	sendingRecords := make([]map[string]interface{}, len(resp.SendingDNSRecords))
	for i, r := range resp.SendingDNSRecords {
		sendingRecords[i] = make(map[string]interface{})
		sendingRecords[i]["id"] = r.Name
		sendingRecords[i]["name"] = r.Name
		sendingRecords[i]["valid"] = r.Valid
		sendingRecords[i]["value"] = r.Value
		sendingRecords[i]["record_type"] = r.RecordType

		if strings.Contains(r.Name, "._domainkey.") {
			sendingRecords[i]["id"] = "_domainkey." + resp.Domain.Name
		}
	}
	d.Set("sending_records", sendingRecords)
	d.Set("sending_records_set", sendingRecords)

	info, err := client.GetDomainTracking(context.Background(), id)
	var openTracking = false
	if info.Open.Active {
		openTracking = true
	}
	d.Set("open_tracking", openTracking)

	return &resp, nil
}
