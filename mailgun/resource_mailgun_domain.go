package mailgun

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mailgun/mailgun-go/v3"
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
			},

			"smtp_login": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"smtp_password": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Computed: true,
			},

			"wildcard": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
			},

			"receiving_records": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
		},
	}
}

func resourceMailgunDomainImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	parts := strings.SplitN(d.Id(), ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		d.Set("region", "us")
	} else {
		d.Set("region", parts[0])
		d.SetId(parts[1])
	}

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

	log.Printf("[DEBUG] Domain create configuration: %#v", opts)

	_, err := client.CreateDomain(context.Background(), name, &opts)

	if err != nil {
		return diag.FromErr(err)
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
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
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
		return diag.FromErr(errc)
	}

	return nil
}

func resourceMailgunDomainRetrieve(id string, client *mailgun.MailgunImpl, d *schema.ResourceData) (*mailgun.DomainResponse, error) {

	resp, err := client.GetDomain(context.Background(), id)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving domain: %s", err)
	}

	d.Set("name", resp.Domain.Name)
	d.Set("smtp_password", resp.Domain.SMTPPassword)
	d.Set("smtp_login", resp.Domain.SMTPLogin)
	d.Set("wildcard", resp.Domain.Wildcard)
	d.Set("spam_action", resp.Domain.SpamAction)

	receivingRecords := make([]map[string]interface{}, len(resp.ReceivingDNSRecords))
	for i, r := range resp.ReceivingDNSRecords {
		receivingRecords[i] = make(map[string]interface{})
		receivingRecords[i]["priority"] = r.Priority
		receivingRecords[i]["valid"] = r.Valid
		receivingRecords[i]["value"] = r.Value
		receivingRecords[i]["record_type"] = r.RecordType
	}
	d.Set("receiving_records", receivingRecords)

	sendingRecords := make([]map[string]interface{}, len(resp.SendingDNSRecords))
	for i, r := range resp.SendingDNSRecords {
		sendingRecords[i] = make(map[string]interface{})
		sendingRecords[i]["name"] = r.Name
		sendingRecords[i]["valid"] = r.Valid
		sendingRecords[i]["value"] = r.Value
		sendingRecords[i]["record_type"] = r.RecordType
	}
	d.Set("sending_records", sendingRecords)

	return &resp, nil
}
