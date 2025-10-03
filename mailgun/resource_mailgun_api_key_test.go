package mailgun

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMailgunApiKey_Basic(t *testing.T) {
	role := "admin"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunApiKeyConfig(role),
				Check: resource.ComposeTestCheckFunc(
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

func testAccCheckMailgunApiKeyConfig(id string) string {
	return `
resource "mailgun_api_key" "foobar" {
    id = "` + id + `"
	description = "Test API key"
	role = "admin"
	kind = "user"
}`
}
