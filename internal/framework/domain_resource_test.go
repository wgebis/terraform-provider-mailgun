package framework_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

func TestAccMailgunDomain_Basic(t *testing.T) {
	var resp mtypes.GetDomainResponse
	id, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", id)
	re := regexp.MustCompile(`^\w+\._domainkey\.` + regexp.QuoteMeta(domain))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6Providers(),
		CheckDestroy:             testAccCheckMailgunDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunDomainConfig(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunDomainExists("mailgun_domain.foobar", &resp),
					testAccCheckMailgunDomainAttributes(domain, &resp),
					resource.TestCheckResourceAttr("mailgun_domain.foobar", "name", domain),
					resource.TestCheckResourceAttr("mailgun_domain.foobar", "spam_action", "disabled"),
					resource.TestCheckResourceAttr("mailgun_domain.foobar", "wildcard", "true"),
					resource.TestCheckResourceAttr("mailgun_domain.foobar", "force_dkim_authority", "true"),
					resource.TestCheckResourceAttr("mailgun_domain.foobar", "open_tracking", "true"),
					resource.TestCheckResourceAttr("mailgun_domain.foobar", "click_tracking", "true"),
					resource.TestCheckResourceAttr("mailgun_domain.foobar", "web_scheme", "https"),
					resource.TestCheckResourceAttr("mailgun_domain.foobar", "use_automatic_sender_security", "true"),
					testAccCheckAnyAttrMatches(
						"mailgun_domain.foobar", "sending_records_set", "name", re),
				),
			},
		},
	})
}

func TestAccMailgunDomain_Import(t *testing.T) {
	resourceName := "mailgun_domain.foobar"
	id, _ := uuid.GenerateUUID()
	domain := fmt.Sprintf("terraform.%s.com", id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6Providers(),
		CheckDestroy:             testAccCheckMailgunDomainDestroy,
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

func testAccCheckMailgunDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_domain" {
			continue
		}
		client, err := mailgunClientFromAttrs(rs.Primary.Attributes)
		if err != nil {
			return err
		}
		resp, err := client.GetDomain(context.Background(), rs.Primary.ID, nil)
		if err == nil {
			return fmt.Errorf("Domain still exists: %#v", resp)
		}
	}
	return nil
}

func testAccCheckMailgunDomainAttributes(domain string, DomainResp *mtypes.GetDomainResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if DomainResp.Domain.Name != domain {
			return fmt.Errorf("Bad name: %s", DomainResp.Domain.Name)
		}
		if DomainResp.Domain.SpamAction != "disabled" {
			return fmt.Errorf("Bad spam_action: %s", DomainResp.Domain.SpamAction)
		}
		if DomainResp.Domain.Wildcard != true {
			return fmt.Errorf("Bad wildcard: %t", DomainResp.Domain.Wildcard)
		}
		if DomainResp.Domain.WebScheme != "https" {
			return fmt.Errorf("Bad web scheme: %s", DomainResp.Domain.WebScheme)
		}
		if DomainResp.Domain.UseAutomaticSenderSecurity != true {
			return fmt.Errorf("Bad use_automatic_sender_security: %t", DomainResp.Domain.UseAutomaticSenderSecurity)
		}
		if len(DomainResp.ReceivingDNSRecords) == 0 || DomainResp.ReceivingDNSRecords[0].Priority == "" {
			return fmt.Errorf("Bad receiving_records: %#v", DomainResp.ReceivingDNSRecords)
		}
		if len(DomainResp.SendingDNSRecords) == 0 || DomainResp.SendingDNSRecords[0].Name == "" {
			return fmt.Errorf("Bad sending_records: %#v", DomainResp.SendingDNSRecords)
		}
		return nil
	}
}

func testAccCheckMailgunDomainExists(n string, DomainResp *mtypes.GetDomainResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Domain ID is set")
		}
		client, err := mailgunClientFromAttrs(rs.Primary.Attributes)
		if err != nil {
			return err
		}
		resp, err := client.GetDomain(context.Background(), rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if resp.Domain.Name != rs.Primary.ID {
			return fmt.Errorf("Domain not found")
		}
		*DomainResp = resp
		return nil
	}
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
	use_automatic_sender_security = true
}`
}
