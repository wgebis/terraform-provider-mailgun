package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

var (
	_ resource.Resource              = (*apiKeyResource)(nil)
	_ resource.ResourceWithConfigure = (*apiKeyResource)(nil)
)

// NewAPIKeyResource is the constructor registered with the framework provider.
func NewAPIKeyResource() resource.Resource {
	return &apiKeyResource{}
}

type apiKeyResource struct {
	cfg *mailgunpkg.Config
}

type apiKeyResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Description    types.String `tfsdk:"description"`
	Kind           types.String `tfsdk:"kind"`
	Region         types.String `tfsdk:"region"`
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

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiresReplaceStr := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: requiresReplaceStr,
			},
			"kind": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString("user"),
				PlanModifiers: requiresReplaceStr,
			},
			"region": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString("us"),
				PlanModifiers: requiresReplaceStr,
			},
			"role": schema.StringAttribute{
				Required:      true,
				PlanModifiers: requiresReplaceStr,
			},
			"domain_name": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: requiresReplaceStr,
			},
			"email": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: requiresReplaceStr,
			},
			"requestor": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: requiresReplaceStr,
			},
			"user_name": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: requiresReplaceStr,
			},
			"expires_at": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"secret": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_disabled": schema.BoolAttribute{
				Computed: true,
			},
			"disabled_reason": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
