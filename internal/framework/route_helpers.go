package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// buildRoutePayload converts the model into a Mailgun mtypes.Route value.
func buildRoutePayload(ctx context.Context, m *routeResourceModel) (mtypes.Route, diag.Diagnostics) {
	var actions []string
	diags := m.Actions.ElementsAs(ctx, &actions, false)
	return mtypes.Route{
		Priority:    int(m.Priority.ValueInt64()),
		Description: m.Description.ValueString(),
		Expression:  m.Expression.ValueString(),
		Actions:     actions,
	}, diags
}

// applyRoute syncs API-returned data back into the model.
func applyRoute(ctx context.Context, m *routeResourceModel, r *mtypes.Route) diag.Diagnostics {
	m.Priority = types.Int64Value(int64(r.Priority))
	m.Description = types.StringValue(r.Description)
	m.Expression = types.StringValue(r.Expression)
	actions, d := types.ListValueFrom(ctx, types.StringType, r.Actions)
	if d.HasError() {
		return d
	}
	m.Actions = actions
	return nil
}
