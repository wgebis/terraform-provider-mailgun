package mailgun

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &routeResource{}
	_ resource.ResourceWithConfigure   = &routeResource{}
	_ resource.ResourceWithImportState = &routeResource{}
)

// NewRouteResource is a helper function to simplify the provider implementation.
func NewRouteResource() resource.Resource {
	return &routeResource{}
}

// routeResource is the resource implementation.
type routeResource struct {
	client *Config
}

// routeResourceModel maps the resource schema data.
type routeResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Priority    types.Int64    `tfsdk:"priority"`
	Region      types.String   `tfsdk:"region"`
	Description types.String   `tfsdk:"description"`
	Expression  types.String   `tfsdk:"expression"`
	Actions     []types.String `tfsdk:"actions"`
}

// Metadata returns the resource type name.
func (r *routeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route"
}

// Schema defines the schema for the resource.
func (r *routeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Mailgun route resource. This can be used to create and manage routes on Mailgun.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the route.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"priority": schema.Int64Attribute{
				Required:    true,
				Description: "The priority of the route (smaller number indicates higher priority).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("us"),
				Description: "The region where route will be created. Valid values are 'us' or 'eu'. Default is 'us'.",
				Validators: []validator.String{
					stringvalidator.OneOf("us", "eu"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the route.",
			},
			"expression": schema.StringAttribute{
				Required:    true,
				Description: "A filter expression to match messages for this route.",
			},
			"actions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "A list of actions to take when a message matches the expression.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *routeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *routeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Mailgun client",
			"Could not create Mailgun client: "+err.Error(),
		)
		return
	}

	opts := mtypes.Route{
		Priority:    int(plan.Priority.ValueInt64()),
		Description: plan.Description.ValueString(),
		Expression:  plan.Expression.ValueString(),
	}

	// Convert []types.String to []string
	actionArray := make([]string, len(plan.Actions))
	for i, action := range plan.Actions {
		actionArray[i] = action.ValueString()
	}
	opts.Actions = actionArray

	log.Printf("[DEBUG] Route create configuration: %v", opts)

	route, err := client.CreateRoute(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating route",
			"Could not create route: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(route.Id)

	log.Printf("[INFO] Route ID: %s", plan.ID.ValueString())

	// Retrieve and update state of route
	err = r.readRoute(ctx, plan.ID.ValueString(), client, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading route",
			"Could not read route after creation: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *routeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Mailgun client",
			"Could not create Mailgun client: "+err.Error(),
		)
		return
	}

	err = r.readRoute(ctx, state.ID.ValueString(), client, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading route",
			"Could not read route: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *routeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan routeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Mailgun client",
			"Could not create Mailgun client: "+err.Error(),
		)
		return
	}

	opts := mtypes.Route{
		Priority:    int(plan.Priority.ValueInt64()),
		Description: plan.Description.ValueString(),
		Expression:  plan.Expression.ValueString(),
	}

	// Convert []types.String to []string
	actionArray := make([]string, len(plan.Actions))
	for i, action := range plan.Actions {
		actionArray[i] = action.ValueString()
	}
	opts.Actions = actionArray

	log.Printf("[DEBUG] Route update configuration: %v", opts)

	route, err := client.UpdateRoute(ctx, plan.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating route",
			"Could not update route: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(route.Id)

	log.Printf("[INFO] Route ID: %s", plan.ID.ValueString())

	// Retrieve and update state of route
	err = r.readRoute(ctx, plan.ID.ValueString(), client, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading route",
			"Could not read route after update: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *routeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Mailgun client",
			"Could not create Mailgun client: "+err.Error(),
		)
		return
	}

	log.Printf("[INFO] Deleting Route: %s", state.ID.ValueString())

	// Destroy the route
	err = client.DeleteRoute(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting route",
			"Could not delete route: "+err.Error(),
		)
		return
	}

	// Give the destroy a chance to take effect - retry for up to 1 minute
	deadline := time.Now().Add(1 * time.Minute)
	for time.Now().Before(deadline) {
		_, err = client.GetRoute(ctx, state.ID.ValueString())
		if err != nil {
			// Route is gone
			log.Printf("[INFO] Route deleted successfully: %s", state.ID.ValueString())
			return
		}
		log.Printf("[INFO] Retrying until route disappears...")
		time.Sleep(5 * time.Second)
	}

	resp.Diagnostics.AddWarning(
		"Route deletion timeout",
		"Route may still exist after timeout",
	)
}

// ImportState imports the resource into Terraform state.
func (r *routeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)

	var region, routeID string
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		region = parts[0]
		routeID = parts[1]
	} else {
		region = "us"
		routeID = req.ID
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), routeID)...)
}

// readRoute reads the route data from Mailgun API and populates the model.
func (r *routeResource) readRoute(ctx context.Context, id string, client *mailgun.Client, model *routeResourceModel) error {
	route, err := client.GetRoute(ctx, id)
	if err != nil {
		return fmt.Errorf("error retrieving route: %w", err)
	}

	model.Priority = types.Int64Value(int64(route.Priority))
	model.Description = types.StringValue(route.Description)
	model.Expression = types.StringValue(route.Expression)

	// Convert []string to []types.String
	actions := make([]types.String, len(route.Actions))
	for i, action := range route.Actions {
		actions[i] = types.StringValue(action)
	}
	model.Actions = actions

	return nil
}
