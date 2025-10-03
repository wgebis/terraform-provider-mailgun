package mailgun

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &apiKeyResource{}
	_ resource.ResourceWithConfigure   = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
)

// NewApiKeyResource is a helper function to simplify the provider implementation.
func NewApiKeyResource() resource.Resource {
	return &apiKeyResource{}
}

// apiKeyResource is the resource implementation.
type apiKeyResource struct {
	client *Config
}

// apiKeyResourceModel maps the resource schema data.
type apiKeyResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Description    types.String `tfsdk:"description"`
	Kind           types.String `tfsdk:"kind"`
	Role           types.String `tfsdk:"role"`
	DomainName     types.String `tfsdk:"domain_name"`
	Email          types.String `tfsdk:"email"`
	Requestor      types.String `tfsdk:"requestor"`
	UserID         types.String `tfsdk:"user_id"`
	UserName       types.String `tfsdk:"user_name"`
	ExpiresAt      types.Int64  `tfsdk:"expires_at"`
	Secret         types.String `tfsdk:"secret"`
	IsDisabled     types.Bool   `tfsdk:"is_disabled"`
	DisabledReason types.String `tfsdk:"disabled_reason"`
}

// Metadata returns the resource type name.
func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

// Schema defines the schema for the resource.
func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Mailgun API key resource. This can be used to create and manage API keys on Mailgun.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kind": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("user"),
				Description: "The kind of API key. Valid values are 'user' or 'sending'. Default is 'user'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "The role for the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_name": schema.StringAttribute{
				Optional:    true,
				Description: "The domain name for the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Description: "The email address for the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"requestor": schema.StringAttribute{
				Computed:    true,
				Description: "The requestor of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Optional:    true,
				Description: "The user ID for the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_name": schema.StringAttribute{
				Optional:    true,
				Description: "The user name for the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expires_at": schema.Int64Attribute{
				Optional:    true,
				Description: "The expiration time of the API key (Unix timestamp).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret value of the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the API key is disabled.",
			},
			"disabled_reason": schema.StringAttribute{
				Computed:    true,
				Description: "The reason why the API key is disabled.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetClient("us")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Mailgun client",
			"Could not create Mailgun client: "+err.Error(),
		)
		return
	}

	opts := mailgun.CreateAPIKeyOptions{}

	if !plan.Description.IsNull() {
		opts.Description = plan.Description.ValueString()
	}
	if !plan.DomainName.IsNull() {
		opts.DomainName = plan.DomainName.ValueString()
	}
	if !plan.Email.IsNull() {
		opts.Email = plan.Email.ValueString()
	}
	if !plan.ExpiresAt.IsNull() {
		opts.Expiration = uint64(plan.ExpiresAt.ValueInt64())
	}
	if !plan.Kind.IsNull() {
		opts.Kind = plan.Kind.ValueString()
	}
	if !plan.UserID.IsNull() {
		opts.UserID = plan.UserID.ValueString()
	}
	if !plan.UserName.IsNull() {
		opts.UserName = plan.UserName.ValueString()
	}

	apiKey, err := client.CreateAPIKey(ctx, plan.Role.ValueString(), &opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating API key",
			"Could not create API key: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(apiKey.ID)
	plan.DisabledReason = types.StringValue(apiKey.DisabledReason)
	plan.IsDisabled = types.BoolValue(apiKey.IsDisabled)
	plan.Requestor = types.StringValue(apiKey.Requestor)
	plan.Secret = types.StringValue(apiKey.Secret)

	log.Printf("[INFO] API key ID: %s", plan.ID.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetClient("us")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Mailgun client",
			"Could not create Mailgun client: "+err.Error(),
		)
		return
	}

	apiKeyList, err := client.ListAPIKeys(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving API key list",
			"Could not retrieve API key list: "+err.Error(),
		)
		return
	}

	var found bool
	for _, key := range apiKeyList {
		if key.ID == state.ID.ValueString() {
			state.DisabledReason = types.StringValue(key.DisabledReason)
			state.IsDisabled = types.BoolValue(key.IsDisabled)
			state.Requestor = types.StringValue(key.Requestor)
			state.Secret = types.StringValue(key.Secret)
			found = true
			break
		}
	}

	if !found {
		log.Printf("[DEBUG] API key not found with ID: %s", state.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// API keys cannot be updated, all fields force replacement
	resp.Diagnostics.AddError(
		"Update not supported",
		"API keys cannot be updated. All changes require resource replacement.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.client.GetClient("us")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Mailgun client",
			"Could not create Mailgun client: "+err.Error(),
		)
		return
	}

	log.Printf("[INFO] Deleting API key: %s", state.ID.ValueString())

	err = client.DeleteAPIKey(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting API key",
			"Could not delete API key: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID directly as the API key ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
