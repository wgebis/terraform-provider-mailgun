package mailgun

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMailgunDomainDataSource_Basic(t *testing.T) {
	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", uuid)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: newProvider(),
		CheckDestroy:      testAccCheckMailgunDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMailgunDomainDataSourceConfig_Basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceMailgunDomainCheck("data.mailgun_domain.test", domain),
				),
			},
		},
	})
}

func testAccDataSourceMailgunDomainCheck(name string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s\n%#v", name, rs)
		}

		attr := rs.Primary.Attributes

		if attr["name"] != domain {
			return fmt.Errorf("bad name %s", attr["name"])
		}

		if attr["spam_action"] != "disabled" {
			return fmt.Errorf("Bad spam_action: %s", attr["spam_action"])
		}

		if attr["wildcard"] != "false" {
			return fmt.Errorf("Bad wildcard: %s", attr["wildcard"])
		}

		if attr["smtp_password"] == "" {
			return fmt.Errorf("Bad smtp_password: %s", attr["smtp_password"])
		}

		return nil
	}
}

func testAccMailgunDomainDataSourceConfig_Basic(domain string) string {
	return fmt.Sprintf(`
resource "mailgun_domain" "foobar" {
	name = "%s"
	spam_action = "disabled"
	smtp_password = "foobarsupersecretpassword"
	wildcard = false
}
data "mailgun_domain" "test" {
	name = mailgun_domain.foobar.id
}
`, domain)
}
