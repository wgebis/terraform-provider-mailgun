package mailgun

import (
	"log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMailgunDomain() *schema.Resource {

	mailgunSchema := resourceMailgunDomain()

	return &schema.Resource{
		Read:   dataSourceMailgunDomainRead,
		Schema: mailgunSchema.Schema,
	}
}

func dataSourceMailgunDomainRead(d *schema.ResourceData, meta interface{}) error {
	region := d.Get("region").(string)
	name := d.Get("name").(string)
	
	log.Printf("[DEBUG] Reading Mailgun domain: name=%s, region=%s", name, region)
	
	client, errc := meta.(*Config).GetClient(region)
	if errc != nil {
		return errc
	}

	_, err := resourceMailgunDomainRetrieve(name, client, d)

	if err != nil {
		return err
	}

	d.SetId(name)
	return nil
}
