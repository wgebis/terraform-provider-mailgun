package mailgun

import (
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
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return errc
	}

	name := d.Get("name").(string)

	_, err := resourceMailgunDomainRetrieve(name, client, d)

	if err != nil {
		return err
	}

	d.SetId(name)
	return nil
}
