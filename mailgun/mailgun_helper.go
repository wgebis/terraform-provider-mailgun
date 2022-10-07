package mailgun

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func setDefaultRegionForImport(d *schema.ResourceData) {
	parts := strings.SplitN(d.Id(), ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		d.Set("region", "us")
	} else {
		d.Set("region", parts[0])
		d.SetId(parts[1])
	}
}
