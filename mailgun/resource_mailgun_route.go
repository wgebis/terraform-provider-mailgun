package mailgun

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mailgun/mailgun-go/v5/mtypes"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mailgun/mailgun-go/v5"
)

func resourceMailgunRoute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMailgunRouteCreate,
		ReadContext:   resourceMailgunRouteRead,
		UpdateContext: resourceMailgunRouteUpdate,
		DeleteContext: resourceMailgunRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMailgunRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"priority": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},

			"region": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "us",
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"expression": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"actions": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceMailgunRouteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	setDefaultRegionForImport(d)

	return []*schema.ResourceData{d}, nil
}

func resourceMailgunRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	opts := mtypes.Route{}

	opts.Priority = d.Get("priority").(int)
	opts.Description = d.Get("description").(string)
	opts.Expression = d.Get("expression").(string)
	actions := d.Get("actions").([]interface{})
	actionArray := []string{}
	for _, i := range actions {
		action := i.(string)
		actionArray = append(actionArray, action)
	}
	opts.Actions = actionArray
	log.Printf("[DEBUG] Route create configuration: %v", opts)

	route, err := client.CreateRoute(ctx, opts)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(route.Id)

	log.Printf("[INFO] Route ID: %s", d.Id())

	// Retrieve and update state of route
	_, err = resourceMailgunRouteRetrieve(ctx, d.Id(), client, d)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMailgunRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	opts := mtypes.Route{}

	opts.Priority = d.Get("priority").(int)
	opts.Description = d.Get("description").(string)
	opts.Expression = d.Get("expression").(string)
	actions := d.Get("actions").([]interface{})
	actionArray := []string{}

	for _, i := range actions {
		action := i.(string)
		actionArray = append(actionArray, action)
	}
	opts.Actions = actionArray

	log.Printf("[DEBUG] Route update configuration: %v", opts)

	route, err := client.UpdateRoute(ctx, d.Id(), opts)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(route.Id)

	log.Printf("[INFO] Route ID: %s", d.Id())

	// Retrieve and update state of route
	_, err = resourceMailgunRouteRetrieve(ctx, d.Id(), client, d)

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMailgunRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	log.Printf("[INFO] Deleting Route: %s", d.Id())

	// Destroy the route
	err := client.DeleteRoute(ctx, d.Id())
	if err != nil {
		return diag.Errorf("Error deleting route: %s", err)
	}

	// Give the destroy a chance to take effect
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err = client.GetRoute(ctx, d.Id())
		if err == nil {
			log.Printf("[INFO] Retrying until route disappears...")
			return resource.RetryableError(
				fmt.Errorf("route seems to still exist; will check again"))
		}
		log.Printf("[INFO] Got error looking for route, seems gone: %s", err)
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceMailgunRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, errc := meta.(*Config).GetClient(d.Get("region").(string))
	if errc != nil {
		return diag.FromErr(errc)
	}

	_, err := resourceMailgunRouteRetrieve(ctx, d.Id(), client, d)

	if err != nil {
		if isNotFound(err) {
			log.Printf("[WARN] Mailgun route %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}

func resourceMailgunRouteRetrieve(ctx context.Context, id string, client *mailgun.Client, d *schema.ResourceData) (*mtypes.Route, error) {

	route, err := client.GetRoute(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("Error retrieving route: %w", err)
	}

	_ = d.Set("priority", route.Priority)
	_ = d.Set("description", route.Description)
	_ = d.Set("expression", route.Expression)
	_ = d.Set("actions", route.Actions)

	return &route, nil
}
