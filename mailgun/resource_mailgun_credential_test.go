package mailgun

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMailgunDomainCredential_Basic(t *testing.T) {

	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraformcred.%s.com", uuid)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunCredentialConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "domain", domain),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "login", "test_crendential"),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "password", "supersecretpassword1234"),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "region", "us"),
				),
			},
		},
	})
}

func TestAccMailgunDomainCredential_Update(t *testing.T) {

	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", uuid)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunCredentialConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "domain", domain),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "login", "test_crendential"),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "password", "supersecretpassword1234"),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "region", "us"),
				),
			},
			{
				Config: testAccCheckMailgunCredentialConfigUpdate(domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "domain", domain),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "login", "test_crendential"),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "password", "azertyuyiop123456987"),
					resource.TestCheckResourceAttr(
						"mailgun_domain_credential.foobar", "region", "us"),
				),
			},
		},
	})
}

func testAccCheckMailgunCredentialConfig(domain string) string {
	return `
resource "mailgun_domain" "foobar" {
    name = "` + domain + `"
	spam_action = "disabled"
	region = "us"
    wildcard = true
}

resource "mailgun_domain_credential" "foobar" {
	domain = mailgun_domain.foobar.id
	login = "test_crendential"
	password = "supersecretpassword1234"
	region = "us"
}`
}

func testAccCheckMailgunCredentialConfigUpdate(domain string) string {
	return `
resource "mailgun_domain" "foobar" {
    name = "` + domain + `"
	spam_action = "disabled"
	region = "us"
    wildcard = true
}

resource "mailgun_domain_credential" "foobar" {
	domain = mailgun_domain.foobar.id
	login = "test_crendential"
	password = "azertyuyiop123456987"
	region = "us"
}`
}
