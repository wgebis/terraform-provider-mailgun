package mailgun

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMailgunWebhook_Basic(t *testing.T) {

	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraformcred.%s.com", uuid)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: newProvider(),
		CheckDestroy:      testAccCheckMailgunWebhookDestroy,
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

func testAccCheckMailgunWebhookDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_webhook" {
			continue
		}

		client, _ := testAccProvider.Meta().(*Config).GetClientForDomain(rs.Primary.Attributes["region"], rs.Primary.Attributes["domain"])

		kind := rs.Primary.Attributes["kind"]
		webhooks, err := client.GetWebhook(context.Background(), kind)

		if err == nil {
			return fmt.Errorf("Webhook still exists: %#v", webhooks)
		}
	}

	return nil
}

func testAccCheckMailgunWebhookConfig(domain string) string {
	return `
resource "mailgun_domain" "foobar" {
    name = "` + domain + `"
	spam_action = "disabled"
	smtp_password = "supersecretpassword1234"
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
