package framework

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

var (
	_ resource.Resource                = (*routeResource)(nil)
	_ resource.ResourceWithImportState = (*routeResource)(nil)
	_ resource.ResourceWithConfigure   = (*routeResource)(nil)
)

// NewRouteResource is the constructor registered with the framework provider.
func NewRouteResource() resource.Resource {
	return &routeResource{}
}

type routeResource struct {
	cfg *mailgunpkg.Config
}

type routeResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Priority    types.Int64  `tfsdk:"priority"`
	Region      types.String `tfsdk:"region"`
	Description types.String `tfsdk:"description"`
	Expression  types.String `tfsdk:"expression"`
	Actions     types.List   `tfsdk:"actions"`
}

func (r *routeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route"
}

func (r *routeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"priority": schema.Int64Attribute{
				Required: true,
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("us"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"expression": schema.StringAttribute{
				Required: true,
			},
			"actions": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *routeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(*mailgunpkg.Config)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data",
			fmt.Sprintf("expected *mailgun.Config, got %T", req.ProviderData))
		return
	}
	r.cfg = cfg
}

// ImportState mirrors the SDKv2 behaviour: bare id defaults region to "us",
// a "region:id" form lets users target a non-default region.
func (r *routeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	region, id := "us", req.ID
	if parts := strings.SplitN(req.ID, ":", 2); len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		region, id = parts[0], parts[1]
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
}

func (r *routeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	opts, d := buildRoutePayload(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Route create configuration: %v", opts)
	created, err := client.CreateRoute(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create route", err.Error())
		return
	}

	plan.ID = types.StringValue(created.Id)
	resp.Diagnostics.Append(applyRoute(ctx, &plan, &created)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	got, err := client.GetRoute(ctx, state.ID.ValueString())
	if err != nil {
		if mailgunpkg.IsNotFound(err) {
			log.Printf("[WARN] Mailgun route %s not found, removing from state", state.ID.ValueString())
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read route", err.Error())
		return
	}

	resp.Diagnostics.Append(applyRoute(ctx, &state, &got)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan routeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	opts, d := buildRoutePayload(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Route update configuration: %v", opts)
	updated, err := client.UpdateRoute(ctx, plan.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update route", err.Error())
		return
	}

	plan.ID = types.StringValue(updated.Id)
	resp.Diagnostics.Append(applyRoute(ctx, &plan, &updated)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	id := state.ID.ValueString()
	log.Printf("[INFO] Deleting Route: %s", id)
	if err := client.DeleteRoute(ctx, id); err != nil {
		resp.Diagnostics.AddError("Failed to delete route", err.Error())
		return
	}

	// Poll until the route disappears (Mailgun is eventually consistent).
	deadline := time.Now().Add(1 * time.Minute)
	for {
		_, err := client.GetRoute(ctx, id)
		if err != nil {
			log.Printf("[INFO] Got error looking for route, seems gone: %s", err)
			return
		}
		if time.Now().After(deadline) {
			resp.Diagnostics.AddError("Timeout waiting for route deletion",
				fmt.Sprintf("route %s still exists after 1 minute", id))
			return
		}
		log.Printf("[INFO] Retrying until route disappears...")
		time.Sleep(2 * time.Second)
	}
}
