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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &credentialResource{}
	_ resource.ResourceWithConfigure   = &credentialResource{}
	_ resource.ResourceWithImportState = &credentialResource{}
)

// NewCredentialResource is a helper function to simplify the provider implementation.
func NewCredentialResource() resource.Resource {
	return &credentialResource{}
}

// credentialResource is the resource implementation.
type credentialResource struct {
	client *Config
}

// credentialResourceModel maps the resource schema data.
type credentialResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Login    types.String `tfsdk:"login"`
	Password types.String `tfsdk:"password"`
	Domain   types.String `tfsdk:"domain"`
	Region   types.String `tfsdk:"region"`
}

// Metadata returns the resource type name.
func (r *credentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

// Schema defines the schema for the resource.
func (r *credentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Mailgun credential resource. This can be used to create and manage SMTP credentials on Mailgun.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the credential (email address).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"login": schema.StringAttribute{
				Required:    true,
				Description: "The login username (without domain).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The password for the credential.",
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain to add the credential to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("us"),
				Description: "The region where credential will be created. Valid values are 'us' or 'eu'. Default is 'us'.",
				Validators: []validator.String{
					stringvalidator.OneOf("us", "eu"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *credentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *credentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan credentialResourceModel
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

	email := fmt.Sprintf("%s@%s", plan.Login.ValueString(), plan.Domain.ValueString())
	password := plan.Password.ValueString()

	log.Printf("[DEBUG] Credential create configuration: email: %s", email)

	err = client.CreateCredential(ctx, plan.Domain.ValueString(), email, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating credential",
			"Could not create credential: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(email)

	log.Printf("[INFO] Create credential ID: %s", plan.ID.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *credentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state credentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.SplitN(state.ID.ValueString(), "@", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid credential ID",
			fmt.Sprintf("The ID of credential '%s' doesn't contain domain!", state.ID.ValueString()),
		)
		return
	}

	login := parts[0]
	domain := parts[1]

	client, err := r.client.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Mailgun client",
			"Could not create Mailgun client: "+err.Error(),
		)
		return
	}

	log.Printf("[DEBUG] Read credential for region '%s' and email '%s'", state.Region.ValueString(), state.ID.ValueString())

	itCredentials := client.ListCredentials(domain, nil)

	ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var page []mtypes.Credential
	found := false

	for itCredentials.Next(ctxTimeout, &page) {
		log.Printf("[DEBUG] Read credential get new page")

		for _, c := range page {
			if c.Login == state.ID.ValueString() {
				state.Login = types.StringValue(login)
				state.Domain = types.StringValue(domain)
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if err := itCredentials.Err(); err != nil {
		resp.Diagnostics.AddError(
			"Error listing credentials",
			"Could not list credentials: "+err.Error(),
		)
		return
	}

	if !found {
		log.Printf("[DEBUG] Credential not found: %s", state.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *credentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan credentialResourceModel
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

	email := fmt.Sprintf("%s@%s", plan.Login.ValueString(), plan.Domain.ValueString())
	password := plan.Password.ValueString()

	log.Printf("[DEBUG] Credential update configuration: email: %s", email)

	err = client.ChangeCredentialPassword(ctx, plan.Domain.ValueString(), email, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating credential",
			"Could not update credential password: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(email)

	log.Printf("[INFO] Update credential ID: %s", plan.ID.ValueString())

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *credentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state credentialResourceModel
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

	email := fmt.Sprintf("%s@%s", state.Login.ValueString(), state.Domain.ValueString())

	log.Printf("[INFO] Deleting credential: %s", email)

	err = client.DeleteCredential(ctx, state.Domain.ValueString(), email)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting credential",
			"Could not delete credential: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *credentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)

	var region, credentialID string
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		region = parts[0]
		credentialID = parts[1]
	} else {
		region = "us"
		credentialID = req.ID
	}

	log.Printf("[DEBUG] Import credential for region '%s' and email '%s'", region, credentialID)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), credentialID)...)
}
