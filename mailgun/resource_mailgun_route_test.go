package mailgun

import (
	"context"
	"fmt"
	"testing"
	"time"

	mailgun "github.com/mailgun/mailgun-go/v3"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccMailgunRoute_Basic(t *testing.T) {
	var route mailgun.Route

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMailgunRouteDestroy,
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

func testAccCheckMailgunRouteDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*mailgun.MailgunImpl)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_route" {
			continue
		}

		ctx := context.Background()

		route, err := (*client).GetRoute(ctx, rs.Primary.ID)

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

		client := testAccProvider.Meta().(*mailgun.MailgunImpl)

		ctx := context.Background()

		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			*Route, err = (*client).GetRoute(ctx, rs.Primary.ID)

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
