package framework

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

var (
	_ resource.Resource                = (*webhookResource)(nil)
	_ resource.ResourceWithImportState = (*webhookResource)(nil)
	_ resource.ResourceWithConfigure   = (*webhookResource)(nil)
)

// NewWebhookResource is the constructor registered with the framework provider.
func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

type webhookResource struct {
	cfg *mailgunpkg.Config
}

type webhookResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Region types.String `tfsdk:"region"`
	Domain types.String `tfsdk:"domain"`
	Kind   types.String `tfsdk:"kind"`
	URLs   types.Set    `tfsdk:"urls"`
}

var allowedWebhookKinds = []string{
	"accepted", "clicked", "complained", "delivered", "opened",
	"permanent_fail", "temporary_fail", "unsubscribed",
}

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("us"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(allowedWebhookKinds...),
				},
			},
			"urls": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState accepts "domain:kind" (region defaults to "us") or
// "region:domain:kind" forms, matching the legacy SDKv2 implementation.
func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 3)
	var region, domain, kind string
	switch len(parts) {
	case 2:
		region, domain, kind = "us", parts[0], parts[1]
	case 3:
		region, domain, kind = parts[0], parts[1], parts[2]
	default:
		resp.Diagnostics.AddError("Invalid import ID",
			"expected 'region:domain:kind' or 'domain:kind'")
		return
	}
	id := fmt.Sprintf("%s:%s:%s", region, domain, kind)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), domain)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kind"), kind)...)
}

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	urls, d := webhookURLs(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := client.CreateWebhook(ctx, plan.Domain.ValueString(), plan.Kind.ValueString(), urls); err != nil {
		resp.Diagnostics.AddError("Failed to create webhook", err.Error())
		return
	}

	plan.ID = types.StringValue(webhookID(&plan))
	log.Printf("[INFO] Create webhook ID: %s", plan.ID.ValueString())

	if d := refreshWebhook(ctx, client, &plan); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	urls, err := client.GetWebhook(ctx, state.Domain.ValueString(), state.Kind.ValueString())
	if err != nil {
		if mailgunpkg.IsNotFound(err) {
			log.Printf("[WARN] Mailgun webhook %s not found, removing from state", state.ID.ValueString())
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read webhook", err.Error())
		return
	}

	urlSet, d := types.SetValueFrom(ctx, types.StringType, urls)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.URLs = urlSet
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	urls, d := webhookURLs(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := client.UpdateWebhook(ctx, plan.Domain.ValueString(), plan.Kind.ValueString(), urls); err != nil {
		resp.Diagnostics.AddError("Failed to update webhook", err.Error())
		return
	}
	log.Printf("[INFO] Update webhook ID: %s", plan.ID.ValueString())

	if d := refreshWebhook(ctx, client, &plan); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	if err := client.DeleteWebhook(ctx, state.Domain.ValueString(), state.Kind.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete webhook", err.Error())
		return
	}
	log.Printf("[INFO] Delete webhook ID: %s", state.ID.ValueString())
}
