package framework_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

func TestAccMailgunRoute_Basic(t *testing.T) {
	var route mtypes.Route

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6Providers(),
		CheckDestroy:             testAccCheckMailgunRouteDestroy,
		Steps: []resource.TestStep{
			{
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6Providers(),
		CheckDestroy:             testAccCheckMailgunRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunRouteConfig,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMailgunRoute_Update(t *testing.T) {
	var route mtypes.Route

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6Providers(),
		CheckDestroy:             testAccCheckMailgunRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunRouteConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunRouteExists("mailgun_route.foobar", &route),
					resource.TestCheckResourceAttr("mailgun_route.foobar", "priority", "0"),
					resource.TestCheckResourceAttr("mailgun_route.foobar", "description", "inbound"),
				),
			},
			{
				Config: testAccCheckMailgunRouteConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunRouteExists("mailgun_route.foobar", &route),
					resource.TestCheckResourceAttr("mailgun_route.foobar", "priority", "10"),
					resource.TestCheckResourceAttr("mailgun_route.foobar", "description", "inbound updated"),
					resource.TestCheckResourceAttr(
						"mailgun_route.foobar", "expression", "match_recipient('.*@updated.example.com')"),
					resource.TestCheckResourceAttr(
						"mailgun_route.foobar", "actions.0", "forward('http://example.com/api/v2/foos/')"),
				),
			},
		},
	})
}

// TestAccMailgunRoute_Recreate covers issue #49: a route deleted out of band
// must trigger a re-create on the next plan instead of failing the Read.
func TestAccMailgunRoute_Recreate(t *testing.T) {
	var route mtypes.Route

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6Providers(),
		CheckDestroy:             testAccCheckMailgunRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckMailgunRouteConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunRouteExists("mailgun_route.foobar", &route),
				),
			},
			{
				PreConfig: func() {
					client, err := mailgunClientFromAttrs(map[string]string{"region": "us"})
					if err != nil {
						t.Fatalf("get client: %s", err)
					}
					if err := client.DeleteRoute(context.Background(), route.Id); err != nil {
						t.Fatalf("delete route out of band: %s", err)
					}
				},
				Config: testAccCheckMailgunRouteConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMailgunRouteExists("mailgun_route.foobar", &route),
				),
			},
		},
	})
}

func testAccCheckMailgunRouteDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mailgun_route" {
			continue
		}
		client, _ := mailgunClientFromAttrs(rs.Primary.Attributes)
		route, err := client.GetRoute(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Route still exists: %#v", route)
		}
	}
	return nil
}

func testAccCheckMailgunRouteExists(n string, route *mtypes.Route) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route ID is set")
		}
		client, errc := mailgunClientFromAttrs(rs.Primary.Attributes)
		if errc != nil {
			return errc
		}
		err := resource.RetryContext(context.Background(), 1*time.Minute, func() *resource.RetryError {
			var err error
			*route, err = client.GetRoute(context.Background(), rs.Primary.ID)
			if err != nil {
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Unable to find Route after retries: %s", err)
		}
		if route.Id != rs.Primary.ID {
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

const testAccCheckMailgunRouteConfigUpdate = `
resource "mailgun_route" "foobar" {
    priority = "10"
    description = "inbound updated"
    expression = "match_recipient('.*@updated.example.com')"
    actions = [
        "forward('http://example.com/api/v2/foos/')",
        "stop()"
    ]
}
`
