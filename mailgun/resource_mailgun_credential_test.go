package mailgun

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mailgun/mailgun-go/v3"
)

func TestAccMailgunCredential_Basic(t *testing.T) {
	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", uuid)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: newProvider(),
		CheckDestroy:      testAccCheckMailgunCrendentialDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckMailgunCredentialConfigWithPassword(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunCredentialExists("mailgun_credential.foobar", t),
					resource.TestCheckResourceAttr(
						"mailgun_credential.foobar", "domain", domain),
					resource.TestCheckResourceAttr(
						"mailgun_credential.foobar", "email", "test@"+domain),
					resource.TestCheckResourceAttr(
						"mailgun_credential.foobar", "password", "supersecretpassword1234"),
					resource.TestCheckResourceAttr(
						"mailgun_credential.foobar", "region", "us"),
				),
			},
		},
	})
}

func testAccCheckMailgunCrendentialDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_credential" {
			continue
		}

		route, err := client.MailgunClient.GetRoute(context.Background(), rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Credential still exists: %#v", route)
		}
	}

	return nil
}

func testAccCheckMailgunCredentialExists(n string, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Credential ID is set")
		}

		client := testAccProvider.Meta().(*Config)

		var itCredentials *mailgun.CredentialsIterator

		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			itCredentials = client.MailgunClient.ListCredentials(nil)

			if itCredentials.TotalCount == 0 {
				return resource.NonRetryableError(fmt.Errorf("No credential found"))
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("Unable to find credential after retries: %s", err)
		}

		var items *[]mailgun.Credential

		for itCredentials.Next(context.Background(), items) {
			t.Log(fmt.Sprintf("%+v\n", itCredentials.Items))
		}

		return nil
	}
}

func testAccCheckMailgunCredentialConfigWithPassword(domain string) string {
	return `resource "mailgun_credential" "foobar" {
	domain = "` + domain + `"
	email = "test@` + domain + `"
	password = "supersecretpassword1234"
	region = "us"
}`
}
