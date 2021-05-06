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
	var credential mailgun.Credential
	randomDomain, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", randomDomain)
	login, _ := uuid.GenerateUUID()
	password, _ := uuid.GenerateUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: newProvider(),
		CheckDestroy:      testAccCheckMailgunCredentialDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckMailgunCredentialConfig(domain, login, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunCredentialExists("mailgun_credential.foobar", &credential),
					resource.TestCheckResourceAttr(
						"mailgun_credential.foobar", "domain", domain),
					resource.TestCheckResourceAttr(
						"mailgun_credential.foobar", "login", login),
					resource.TestCheckResourceAttr(
						"mailgun_credential.foobar", "password", password),
				),
			},
		},
	})
}

func TestAccMailgunCredential_Import(t *testing.T) {
	resourceName := "mailgun_credential.foobar"
	randomDomain, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", randomDomain)
	login, _ := uuid.GenerateUUID()
	password, _ := uuid.GenerateUUID()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: newProvider(),
		CheckDestroy:      testAccCheckMailgunCredentialDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckMailgunCredentialConfig(domain, login, password),
			},

			resource.TestStep{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckMailgunCredentialDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_credential" {
			continue
		}

		client, errc := testAccProvider.Meta().(*Config).GetClientForDomain("", rs.Primary.Attributes["domain"])
		if errc != nil {
			return errc
		}

		it := client.ListCredentials(nil)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		var page []mailgun.Credential
		for it.Next(ctx, &page) {
			for _, item := range page {
				if item.Login == rs.Primary.Attributes["login"] {
					return fmt.Errorf("Credential %s still exists", item.Login)
				}
			}
		}
	}

	return nil
}

func testAccCheckMailgunCredentialExists(n string, credential *mailgun.Credential) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No credential ID is set")
		}

		client, errc := testAccProvider.Meta().(*Config).GetClientForDomain("", rs.Primary.Attributes["domain"])
		if errc != nil {
			return errc
		}

		it := client.ListCredentials(nil)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		var page []mailgun.Credential
		for it.Next(ctx, &page) {
			for _, item := range page {
				if item.Login == rs.Primary.Attributes["login"] {
					*credential = item
					return nil
				}
			}
		}

		return fmt.Errorf("Credential not found")
	}
}

func testAccCheckMailgunCredentialConfig(domain string, login string, password string) string {
	return `resource "mailgun_credential" "foobar" {
    domain = "` + domain + `"
	login = "` + login + `"
	password = "` + password + `"
}`
}
