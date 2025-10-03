package mailgun

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMailgunDomain_Basic(t *testing.T) {
	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", uuid)
	re := regexp.MustCompile(`^\w+\._domainkey\.` + regexp.QuoteMeta(domain))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunDomainConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"mailgun_domain.foobar", "name", domain),
					resource.TestCheckResourceAttr(
						"mailgun_domain.foobar", "spam_action", "disabled"),
					resource.TestCheckResourceAttr(
						"mailgun_domain.foobar", "wildcard", "true"),
					resource.TestCheckResourceAttr(
						"mailgun_domain.foobar", "force_dkim_authority", "true"),
					resource.TestCheckResourceAttr(
						"mailgun_domain.foobar", "open_tracking", "true"),
					resource.TestCheckResourceAttr(
						"mailgun_domain.foobar", "click_tracking", "true"),
					resource.TestCheckResourceAttr(
						"mailgun_domain.foobar", "receiving_records.0.priority", "10"),
					resource.TestCheckResourceAttr(
						"mailgun_domain.foobar", "web_scheme", "https"),
					resource.TestCheckResourceAttr(
						"mailgun_domain.foobar", "receiving_records.0.priority", "10"),
					resource.TestMatchResourceAttr(
						"mailgun_domain.foobar", "sending_records.0.name", re),
					resource.TestMatchResourceAttr(
						"mailgun_domain.foobar", "sending_records_set.0.name", re),
				),
			},
		},
	})
}

func TestAccMailgunDomain_Import(t *testing.T) {
	resourceName := "mailgun_domain.foobar"
	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", uuid)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunDomainConfig(domain),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dkim_key_size", "force_dkim_authority"},
			},
		},
	})
}

func testAccCheckMailgunDomainConfig(domain string) string {
	return `
resource "mailgun_domain" "foobar" {
    name = "` + domain + `"
	spam_action = "disabled"
	region = "us"
    wildcard = true
	force_dkim_authority = true
	open_tracking = true
	click_tracking = true
	web_scheme = "https"
}`
}
