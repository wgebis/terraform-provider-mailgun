package framework

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	opts := mailgun.CreateAPIKeyOptions{
		Description: plan.Description.ValueString(),
		DomainName:  plan.DomainName.ValueString(),
		Email:       plan.Email.ValueString(),
		Expiration:  uint64(plan.ExpiresAt.ValueInt64()),
		Kind:        plan.Kind.ValueString(),
		UserID:      plan.UserID.ValueString(),
		UserName:    plan.UserName.ValueString(),
	}

	apiKey, err := client.CreateAPIKey(ctx, plan.Role.ValueString(), &opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API key", err.Error())
		return
	}

	plan.ID = types.StringValue(apiKey.ID)
	plan.Requestor = types.StringValue(apiKey.Requestor)
	plan.Secret = types.StringValue(apiKey.Secret)
	plan.IsDisabled = types.BoolValue(apiKey.IsDisabled)
	plan.DisabledReason = types.StringValue(apiKey.DisabledReason)
	log.Printf("[INFO] API key ID: %s", plan.ID.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	apiKey, found, err := lookupAPIKey(ctx, client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to retrieve API key list", err.Error())
		return
	}
	if !found {
		log.Printf("[DEBUG] API key not found with ID: %s", state.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}

	applyAPIKeyToModel(&state, apiKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is unreachable: every writable attribute uses RequiresReplace, so
// the framework always plans a destroy/create instead of an in-place update.
// The method exists only to satisfy resource.Resource.
func (r *apiKeyResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update unsupported",
		"mailgun_api_key has no updatable attributes; this should not be reachable")
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	log.Printf("[INFO] Deleting API key: %s", state.ID.ValueString())
	if err := client.DeleteAPIKey(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete API key", err.Error())
		return
	}
}

// lookupAPIKey returns the API key with the given ID and a found flag. It
// pages through ListAPIKeys because the Mailgun API has no GET-by-id endpoint.
func lookupAPIKey(ctx context.Context, client *mailgun.Client, id string) (mtypes.APIKey, bool, error) {
	keys, err := client.ListAPIKeys(ctx, nil)
	if err != nil {
		return mtypes.APIKey{}, false, err
	}
	for _, k := range keys {
		if k.ID == id {
			return k, true, nil
		}
	}
	return mtypes.APIKey{}, false, nil
}

// applyAPIKeyToModel mirrors the legacy applyAPIKey: it copies fields from
// the API response into the model and preserves the existing secret when the
// API returns an empty value (regression #73). Optional string attributes
// that the API returns as "" are mapped to types.StringNull so they don't
// drift against null values from HCL — otherwise framework would plan a
// force-replace on ForceNew fields like domain_name/user_name.
func applyAPIKeyToModel(m *apiKeyResourceModel, k mtypes.APIKey) {
	m.Description = stringOrNull(k.Description)
	m.Kind = types.StringValue(k.Kind)
	m.Role = types.StringValue(k.Role)
	m.DomainName = stringOrNull(k.DomainName)
	m.UserName = stringOrNull(k.UserName)
	m.Requestor = types.StringValue(k.Requestor)
	m.IsDisabled = types.BoolValue(k.IsDisabled)
	m.DisabledReason = types.StringValue(k.DisabledReason)
	if k.Secret != "" {
		m.Secret = types.StringValue(k.Secret)
	}
}

// stringOrNull returns types.StringNull for "" and types.StringValue otherwise.
// Used to bridge the SDKv2 (empty == null) vs framework (distinct) gap when
// reading API responses where a missing optional field comes back as "".
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
