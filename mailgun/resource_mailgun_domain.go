package mailgun

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	mailgun "github.com/mailgun/mailgun-go/v3"
)

func resourceMailgunDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceMailgunDomainCreate,
		Read:   resourceMailgunDomainRead,
		Delete: resourceMailgunDomainDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"spam_action": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
				Optional: true,
			},

			"smtp_login": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"wildcard": {
				Type:     schema.TypeBool,
				Computed: true,
				ForceNew: true,
				Optional: true,
			},

			"receiving_records": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"priority": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"record_type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"valid": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": &schema.Schema{
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
						"record_type": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"valid": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceMailgunDomainCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(mailgun.Mailgun)

	opts := mailgun.CreateDomainOptions{}

	name := d.Get("name").(string)

	// opts.Password = d.Get("smtp_password").(string)
	opts.SpamAction = mailgun.SpamAction(d.Get("spam_action").(string))
	opts.Wildcard = d.Get("wildcard").(bool)

	log.Printf("[DEBUG] Domain create configuration: %#v", opts)

	ctx := context.Background()

	_, err := client.CreateDomain(ctx, name, &opts)

	if err != nil {
		return err
	}

	d.SetId(name)

	log.Printf("[INFO] Domain ID: %s", d.Id())

	// Retrieve and update state of domain
	_, err = resourceMailgunDomainRetrieve(d.Id(), client, d)

	if err != nil {
		return err
	}

	return nil
}

func resourceMailgunDomainDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(mailgun.Mailgun)

	log.Printf("[INFO] Deleting Domain: %s", d.Id())

	ctx := context.Background()

	// Destroy the domain
	err := client.DeleteDomain(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting domain: %s", err)
	}

	// Give the destroy a chance to take effect
	return resource.Retry(1*time.Minute, func() *resource.RetryError {
		ctx := context.Background()

		_, err = client.GetDomain(ctx, d.Id())
		if err == nil {
			log.Printf("[INFO] Retrying until domain disappears...")
			return resource.RetryableError(
				fmt.Errorf("Domain seems to still exist; will check again."))
		}
		log.Printf("[INFO] Got error looking for domain, seems gone: %s", err)
		return nil
	})
}

func resourceMailgunDomainRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(mailgun.Mailgun)

	_, err := resourceMailgunDomainRetrieve(d.Id(), client, d)

	if err != nil {
		return err
	}

	return nil
}

func resourceMailgunDomainRetrieve(id string, client mailgun.Mailgun, d *schema.ResourceData) (*mailgun.DomainResponse, error) {

	ctx := context.Background()

	resp, err := client.GetDomain(ctx, id)

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
