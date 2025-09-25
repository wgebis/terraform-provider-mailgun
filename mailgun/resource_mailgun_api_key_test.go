package mailgun

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

func TestAccMailgunApiKey_Basic(t *testing.T) {
	var resp mtypes.APIKey
	role := "admin"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: newProvider(),
		CheckDestroy:      testAccCheckMailgunApiKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunApiKeyConfig(role),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunApiKeyExists("mailgun_api_key.foobar", &resp),
					testAccCheckMailgunApiKeyAttributes(role, &resp),
					resource.TestCheckResourceAttr(
						"mailgun_api_key.foobar", "role", role),
					resource.TestCheckResourceAttr(
						"mailgun_api_key.foobar", "description", "Test API key"),
					resource.TestCheckResourceAttr(
						"mailgun_api_key.foobar", "kind", "user"),
				),
			},
		},
	})
}

func testAccCheckMailgunApiKeyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_api_key" {
			continue
		}

		client, errc := testAccProvider.Meta().(*Config).GetClient(rs.Primary.Attributes["region"])
		if errc != nil {
			return errc
		}

		resp, _ := client.ListAPIKeys(context.Background(), nil)

		for _, key := range resp {
			if key.ID == rs.Primary.ID {
				return fmt.Errorf("API key still exists: %#v", resp)
			}
		}
	}

	return nil
}

func testAccCheckMailgunApiKeyAttributes(id string, APIKey *mtypes.APIKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if APIKey.ID != id {
			return fmt.Errorf("Bad ID: %s", APIKey.ID)
		}

		return nil
	}
}

func testAccCheckMailgunApiKeyExists(n string, APIKey *mtypes.APIKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API key ID is set")
		}

		client, errc := testAccProvider.Meta().(*Config).GetClient(rs.Primary.Attributes["region"])
		if errc != nil {
			return errc
		}

		resp, err := client.ListAPIKeys(context.Background(), nil)

		if err != nil {
			return err
		}

		for _, key := range resp {
			if key.ID == rs.Primary.ID {
				*APIKey = key
				return nil
			}
		}

		return fmt.Errorf("API key not found")
	}
}

func testAccCheckMailgunApiKeyConfig(id string) string {
	return `
resource "mailgun_api_key" "foobar" {
    id = "` + id + `"
	description = "Test API key"
	role = "admin"
	kind = "user"
}`
}
