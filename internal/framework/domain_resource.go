package framework

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

var (
	_ resource.Resource                 = (*domainResource)(nil)
	_ resource.ResourceWithImportState  = (*domainResource)(nil)
	_ resource.ResourceWithModifyPlan   = (*domainResource)(nil)
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

// ModifyPlan pre-populates *_records_set during create/replace so the plan
// shows users the predictable DNS records they will need to add. This mirrors
// the legacy CustomizeDiff behaviour.
func (r *domainResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan, state domainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	nameChanged := req.State.Raw.IsNull()
	if !nameChanged {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		nameChanged = !plan.Name.Equal(state.Name)
	}
	if !nameChanged || plan.Name.IsUnknown() || plan.Name.IsNull() {
		return
	}

	name := plan.Name.ValueString()
	sending := []sendingRecordModel{
		{ID: types.StringValue(name)},
		{ID: types.StringValue("_domainkey." + name)},
		{ID: types.StringValue("email." + name)},
	}
	for i := range sending {
		sending[i].Name = types.StringUnknown()
		sending[i].RecordType = types.StringUnknown()
		sending[i].Valid = types.StringUnknown()
		sending[i].Value = types.StringUnknown()
	}
	sendingSet, d := types.SetValueFrom(ctx, sendingRecordObjectType(), sending)
	resp.Diagnostics.Append(d...)

	receiving := []receivingRecordModel{
		{ID: types.StringValue("mxa.mailgun.org")},
		{ID: types.StringValue("mxb.mailgun.org")},
	}
	for i := range receiving {
		receiving[i].Priority = types.StringUnknown()
		receiving[i].RecordType = types.StringUnknown()
		receiving[i].Valid = types.StringUnknown()
		receiving[i].Value = types.StringUnknown()
	}
	receivingSet, d := types.SetValueFrom(ctx, receivingRecordObjectType(), receiving)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("sending_records_set"), sendingSet)...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("receiving_records_set"), receivingSet)...)
}

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
