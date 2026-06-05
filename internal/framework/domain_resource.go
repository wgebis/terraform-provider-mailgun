package framework

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

var (
	_ resource.Resource                 = (*domainResource)(nil)
	_ resource.ResourceWithImportState  = (*domainResource)(nil)
	_ resource.ResourceWithConfigure    = (*domainResource)(nil)
	_ resource.ResourceWithUpgradeState = (*domainResource)(nil)
)

// NewDomainResource is the constructor registered with the framework provider.
func NewDomainResource() resource.Resource {
	return &domainResource{}
}

type domainResource struct {
	cfg *mailgunpkg.Config
}

func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = domainResourceSchema()
}

func (r *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// No ModifyPlan: the previous implementation tried to pre-populate
// sending_records_set / receiving_records_set during create/replace so users
// would see predictable DNS record ids in the plan. The prediction was
// inherently unreliable — the DKIM record id depends on the Mailgun-default
// selector when dkim_selector is not set, and the API can return a different
// number of records than predicted (e.g. tracking entries). Both led to
// "planned set element does not correlate" / "length changed" errors after
// apply. Computed + setplanmodifier.UseStateForUnknown on the schema is
// sufficient to keep these stable across refreshes; on first create the plan
// simply shows "(known after apply)".

// UpgradeState drops the deprecated sending_records / receiving_records
// TypeList attributes from state created by SDKv2 (schema version 0).
func (r *domainResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: domainSchemaV0(),
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior domainResourceModelV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}
				upgraded := prior.toV1()
				resp.Diagnostics.Append(resp.State.Set(ctx, upgraded)...)
			},
		},
	}
}

// ImportState supports two id formats: "name" (defaults region to "us") and
// "region:name" (matching the SDKv2 helper).
func (r *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	region, name := "us", req.ID
	if parts := strings.SplitN(req.ID, ":", 2); len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		region, name = parts[0], parts[1]
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
}
