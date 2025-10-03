package mailgun

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
	"github.com/mailgun/mailgun-go/v5"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &webhookResource{}
	_ resource.ResourceWithConfigure   = &webhookResource{}
	_ resource.ResourceWithImportState = &webhookResource{}
)

// NewWebhookResource is a helper function to simplify the provider implementation.
func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

// webhookResource is the resource implementation.
type webhookResource struct {
	client *Config
}

// webhookResourceModel maps the resource schema data.
type webhookResourceModel struct {
	ID     types.String   `tfsdk:"id"`
	Region types.String   `tfsdk:"region"`
	Domain types.String   `tfsdk:"domain"`
	Kind   types.String   `tfsdk:"kind"`
	URLs   []types.String `tfsdk:"urls"`
}

// Metadata returns the resource type name.
func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

// Schema defines the schema for the resource.
func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Mailgun webhook resource. This can be used to create and manage webhooks on Mailgun.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the webhook (format: region:domain:kind).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("us"),
				Description: "The region where webhook will be created. Valid values are 'us' or 'eu'. Default is 'us'.",
				Validators: []validator.String{
					stringvalidator.OneOf("us", "eu"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain to add the webhook to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": schema.StringAttribute{
				Required:    true,
				Description: "The kind of webhook. Valid values are 'accepted', 'clicked', 'complained', 'delivered', 'opened', 'permanent_fail', 'temporary_fail', or 'unsubscribed'.",
				Validators: []validator.String{
					stringvalidator.OneOf("accepted", "clicked", "complained", "delivered", "opened", "permanent_fail", "temporary_fail", "unsubscribed"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"urls": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The URLs to send webhook requests to.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookResourceModel
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

	kind := plan.Kind.ValueString()

	// Convert []types.String to []string
	stringUrls := make([]string, len(plan.URLs))
	for i, url := range plan.URLs {
		stringUrls[i] = url.ValueString()
	}

	err = client.CreateWebhook(ctx, plan.Domain.ValueString(), kind, stringUrls)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating webhook",
			"Could not create webhook: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(r.generateID(plan.Region.ValueString(), plan.Domain.ValueString(), kind))

	log.Printf("[INFO] Create webhook ID: %s", plan.ID.ValueString())

	// Read back the webhook to ensure state is synced
	err = r.readWebhook(ctx, client, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading webhook",
			"Could not read webhook after creation: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookResourceModel
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

	err = r.readWebhook(ctx, client, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading webhook",
			"Could not read webhook: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookResourceModel
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

	kind := plan.Kind.ValueString()

	// Convert []types.String to []string
	stringUrls := make([]string, len(plan.URLs))
	for i, url := range plan.URLs {
		stringUrls[i] = url.ValueString()
	}

	err = client.UpdateWebhook(ctx, plan.Domain.ValueString(), kind, stringUrls)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating webhook",
			"Could not update webhook: "+err.Error(),
		)
		return
	}

	log.Printf("[INFO] Update webhook ID: %s", plan.ID.ValueString())

	// Read back the webhook to ensure state is synced
	err = r.readWebhook(ctx, client, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading webhook",
			"Could not read webhook after update: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookResourceModel
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

	kind := state.Kind.ValueString()

	log.Printf("[INFO] Delete webhook ID: %s", state.ID.ValueString())

	err = client.DeleteWebhook(ctx, state.Domain.ValueString(), kind)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting webhook",
			"Could not delete webhook: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)

	var region string
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		region = parts[0]
	} else {
		region = "us"
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// readWebhook reads the webhook data from Mailgun API and populates the model.
func (r *webhookResource) readWebhook(ctx context.Context, client *mailgun.Client, model *webhookResourceModel) error {
	kind := model.Kind.ValueString()
	urls, err := client.GetWebhook(ctx, model.Domain.ValueString(), kind)
	if err != nil {
		return fmt.Errorf("error retrieving webhook: %w", err)
	}

	// Convert []string to []types.String
	urlsList := make([]types.String, len(urls))
	for i, url := range urls {
		urlsList[i] = types.StringValue(url)
	}
	model.URLs = urlsList

	return nil
}

// generateID generates a unique ID for the webhook.
func (r *webhookResource) generateID(region, domain, kind string) string {
	return fmt.Sprintf("%s:%s:%s", region, domain, kind)
}
