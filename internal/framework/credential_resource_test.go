package framework_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

func TestAccMailgunDomainCredential_Basic(t *testing.T) {
	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraformcred.%s.com", uuid)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6Providers(),
		CheckDestroy:             testAccCheckMailgunCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunCredentialConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunCredentialExists("mailgun_domain_credential.foobar"),
					resource.TestCheckResourceAttr("mailgun_domain_credential.foobar", "domain", domain),
					resource.TestCheckResourceAttr("mailgun_domain_credential.foobar", "login", "test_crendential"),
					resource.TestCheckResourceAttr("mailgun_domain_credential.foobar", "password", "supersecretpassword1234"),
					resource.TestCheckResourceAttr("mailgun_domain_credential.foobar", "region", "us"),
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
		ProtoV6ProviderFactories: protoV6Providers(),
		CheckDestroy:             testAccCheckMailgunCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunCredentialConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunCredentialExists("mailgun_domain_credential.foobar"),
					resource.TestCheckResourceAttr("mailgun_domain_credential.foobar", "password", "supersecretpassword1234"),
				),
			},
			{
				Config: testAccCheckMailgunCredentialConfigUpdate(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunCredentialExists("mailgun_domain_credential.foobar"),
					resource.TestCheckResourceAttr("mailgun_domain_credential.foobar", "password", "azertyuyiop123456987"),
				),
			},
		},
	})
}

func TestAccMailgunDomainCredential_Import(t *testing.T) {
	resourceName := "mailgun_domain_credential.foobar"
	uuid, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraformcredimp.%s.com", uuid)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6Providers(),
		CheckDestroy:             testAccCheckMailgunCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunCredentialConfig(domain),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccCheckMailgunCredentialDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_domain_credential" {
			continue
		}
		client, err := mailgunClientFromAttrs(rs.Primary.Attributes)
		if err != nil {
			return err
		}
		resp, err := client.GetDomain(context.Background(), rs.Primary.Attributes["domain"], nil)
		if err == nil {
			it := client.ListCredentials(rs.Primary.Attributes["domain"], nil)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()
			var page []mtypes.Credential
			for it.Next(ctx, &page) {
				for _, c := range page {
					if c.Login == rs.Primary.ID {
						return fmt.Errorf("The credential '%s' found! Created at: %s", rs.Primary.ID, c.CreatedAt.String())
					}
				}
			}
			if err := it.Err(); err != nil {
				return err
			}
			return fmt.Errorf("Domain still exists: %#v", resp)
		}
	}
	return nil
}

func testAccCheckMailgunCredentialExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No domain credential ID is set")
		}
		client, _ := mailgunClientFromAttrs(rs.Primary.Attributes)
		it := client.ListCredentials(rs.Primary.Attributes["domain"], nil)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()
		var page []mtypes.Credential
		for it.Next(ctx, &page) {
			for _, c := range page {
				if c.Login == rs.Primary.ID {
					return nil
				}
			}
		}
		if err := it.Err(); err != nil {
			return err
		}
		return fmt.Errorf("The credential '%s' not found!", rs.Primary.ID)
	}
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
