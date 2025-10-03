package mailgun

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMailgunWebhook_Basic(t *testing.T) {

	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraformcred.%s.com", uuid)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunWebhookConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"mailgun_webhook.foobar", "domain", domain),
					resource.TestCheckResourceAttr(
						"mailgun_webhook.foobar", "region", "us"),
					resource.TestCheckResourceAttr(
						"mailgun_webhook.foobar", "kind", "delivered"),
					resource.TestCheckResourceAttr(
						"mailgun_webhook.foobar", "urls.0", "https://hoge.com"),
				),
			},
		},
	})
}

func TestAccMailgunWebhook_Import(t *testing.T) {
	resourceName := "mailgun_webhook.foobar"
	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", uuid)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunWebhookConfig(domain),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMailgunWebhookConfig(domain string) string {
	return `
resource "mailgun_domain" "foobar" {
    name = "` + domain + `"
	spam_action = "disabled"
	region = "us"
    wildcard = true
}

resource "mailgun_webhook" "foobar" {
  domain = mailgun_domain.foobar.id
  region = "us"
  kind = "delivered"
  urls = ["https://hoge.com"]
}`
}
