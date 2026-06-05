package framework

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

var (
	_ resource.Resource                = (*credentialResource)(nil)
	_ resource.ResourceWithImportState = (*credentialResource)(nil)
	_ resource.ResourceWithConfigure   = (*credentialResource)(nil)
)

// NewCredentialResource is the constructor registered with the framework
// provider for mailgun_domain_credential.
func NewCredentialResource() resource.Resource {
	return &credentialResource{}
}

type credentialResource struct {
	cfg *mailgunpkg.Config
}

type credentialResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Login    types.String `tfsdk:"login"`
	Password types.String `tfsdk:"password"`
	Domain   types.String `tfsdk:"domain"`
	Region   types.String `tfsdk:"region"`
}

func (r *credentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_credential"
}

func (r *credentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"login": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"domain": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
		},
	}
}

func (r *credentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState parses the SDKv2-compatible "region:login@domain" or bare
// "login@domain" form. Region defaults to "us" when omitted.
func (r *credentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	region, id := "us", req.ID
	if at := splitRegion(req.ID); at != "" {
		region, id = at, req.ID[len(at)+1:]
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
}

func (r *credentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan credentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	domain := plan.Domain.ValueString()
	email := fmt.Sprintf("%s@%s", plan.Login.ValueString(), domain)
	log.Printf("[DEBUG] Credential create configuration: email: %s", email)

	if err := client.CreateCredential(ctx, domain, email, plan.Password.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to create credential", err.Error())
		return
	}

	plan.ID = types.StringValue(email)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *credentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state credentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	if domain == "" {
		// imported state: derive domain from id (login@domain).
		if at := lastIndexAt(state.ID.ValueString()); at != -1 {
			domain = state.ID.ValueString()[at+1:]
			state.Domain = types.StringValue(domain)
			state.Login = types.StringValue(state.ID.ValueString()[:at])
		}
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	found, err := credentialExists(ctx, client, domain, state.ID.ValueString())
	if err != nil {
		if mailgunpkg.IsNotFound(err) {
			log.Printf("[WARN] Mailgun credential %s not found, removing from state", state.ID.ValueString())
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read credential", err.Error())
		return
	}
	if !found {
		log.Printf("[WARN] Mailgun credential %s not found, removing from state", state.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *credentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan credentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	domain := plan.Domain.ValueString()
	email := fmt.Sprintf("%s@%s", plan.Login.ValueString(), domain)
	log.Printf("[DEBUG] Credential update configuration: email: %s", email)

	if err := client.ChangeCredentialPassword(ctx, domain, email, plan.Password.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to update credential", err.Error())
		return
	}

	plan.ID = types.StringValue(email)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *credentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state credentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	if err := client.DeleteCredential(ctx, state.Domain.ValueString(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete credential", err.Error())
		return
	}
}

// credentialExists reports whether a credential with the given email lives on
// the domain. Pages through ListCredentials with a 30s timeout to match the
// legacy behaviour.
func credentialExists(ctx context.Context, client *mailgun.Client, domain, email string) (bool, error) {
	it := client.ListCredentials(domain, nil)
	pageCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var page []mtypes.Credential
	for it.Next(pageCtx, &page) {
		for _, c := range page {
			if c.Login == email {
				return true, nil
			}
		}
	}
	return false, it.Err()
}

// splitRegion returns the leading "region" of an "region:rest" id, or "" if
// the id is not in that form. Distinguishes from "login@domain" by requiring
// the colon to come before any "@".
func splitRegion(id string) string {
	for i := 0; i < len(id); i++ {
		switch id[i] {
		case ':':
			return id[:i]
		case '@':
			return ""
		}
	}
	return ""
}

func lastIndexAt(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '@' {
			return i
		}
	}
	return -1
}
