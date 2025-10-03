package mailgun

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMailgunDomainDataSource_Basic(t *testing.T) {
	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", uuid)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMailgunDomainDataSourceConfig_Basic(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_domain.test", "name", domain),
					resource.TestCheckResourceAttr("data.mailgun_domain.test", "spam_action", "disabled"),
					resource.TestCheckResourceAttr("data.mailgun_domain.test", "wildcard", "false"),
				),
			},
		},
	})
}

func testAccMailgunDomainDataSourceConfig_Basic(domain string) string {
	return fmt.Sprintf(`
resource "mailgun_domain" "foobar" {
	name = "%s"
	spam_action = "disabled"
	wildcard = false
}

data "mailgun_domain" "test" {
	name = mailgun_domain.foobar.id
	region = mailgun_domain.foobar.region
}
`, domain)
}
