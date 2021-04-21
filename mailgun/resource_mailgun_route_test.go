package mailgun

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mailgun/mailgun-go/v4"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMailgunRoute_Basic(t *testing.T) {
	var route mailgun.Route

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: newProvider(),
		CheckDestroy:      testAccCheckMailgunRouteDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckMailgunRouteConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunRouteExists("mailgun_route.foobar", &route),
					resource.TestCheckResourceAttr(
						"mailgun_route.foobar", "priority", "0"),
					resource.TestCheckResourceAttr(
						"mailgun_route.foobar", "description", "inbound"),
					resource.TestCheckResourceAttr(
						"mailgun_route.foobar", "expression", "match_recipient('.*@example.com')"),
					resource.TestCheckResourceAttr(
						"mailgun_route.foobar", "actions.0", "forward('http://example.com/api/v1/foos/')"),
					resource.TestCheckResourceAttr(
						"mailgun_route.foobar", "actions.1", "stop()"),
				),
			},
		},
	})
}

func TestAccMailgunRoute_Import(t *testing.T) {
	resourceName := "mailgun_route.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: newProvider(),
		CheckDestroy:      testAccCheckMailgunRouteDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckMailgunRouteConfig,
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckMailgunRouteDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_route" {
			continue
		}

		route, err := client.MailgunClient.GetRoute(context.Background(), rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Route still exists: %#v", route)
		}
	}

	return nil
}

func testAccCheckMailgunRouteExists(n string, Route *mailgun.Route) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route ID is set")
		}

		client := testAccProvider.Meta().(*Config)

		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			*Route, err = client.MailgunClient.GetRoute(context.Background(), rs.Primary.ID)

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("Unable to find Route after retries: %s", err)
		}

		if Route.Id != rs.Primary.ID {
			return fmt.Errorf("Route not found")
		}

		return nil
	}
}

const testAccCheckMailgunRouteConfig = `
resource "mailgun_route" "foobar" {
    priority = "0"
    description = "inbound"
    expression = "match_recipient('.*@example.com')"
    actions = [
        "forward('http://example.com/api/v1/foos/')",
        "stop()"
    ]
}
`
