package framework

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	name := plan.Name.ValueString()
	opts := mailgun.CreateDomainOptions{
		SpamAction:                 mtypes.SpamAction(plan.SpamAction.ValueString()),
		Password:                   plan.SmtpPassword.ValueString(),
		Wildcard:                   plan.Wildcard.ValueBool(),
		DKIMKeySize:                int(plan.DkimKeySize.ValueInt64()),
		ForceDKIMAuthority:         plan.ForceDkimAuthority.ValueBool(),
		UseAutomaticSenderSecurity: plan.UseAutomaticSenderSecurity.ValueBool(),
		WebScheme:                  plan.WebScheme.ValueString(),
	}

	log.Printf("[DEBUG] Domain create configuration: %#v", opts)
	createResp, err := client.CreateDomain(ctx, name, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create domain", err.Error())
		return
	}

	if sel := plan.DkimSelector.ValueString(); sel != "" {
		if err := client.UpdateDomainDkimSelector(ctx, name, sel); err != nil {
			resp.Diagnostics.AddError("Failed to set DKIM selector", err.Error())
			return
		}
	}
	if plan.OpenTracking.ValueBool() {
		if err := client.UpdateOpenTracking(ctx, name, "yes"); err != nil {
			resp.Diagnostics.AddError("Failed to enable open tracking", err.Error())
			return
		}
	}
	if plan.ClickTracking.ValueBool() {
		if err := client.UpdateClickTracking(ctx, name, "yes"); err != nil {
			resp.Diagnostics.AddError("Failed to enable click tracking", err.Error())
			return
		}
	}

	plan.ID = types.StringValue(name)
	planPwd := plan.SmtpPassword
	diags, _ := refreshDomain(ctx, client, name, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Mailgun never returns smtp_password from GetDomain, so apply precedence:
	// 1) value from the user's plan if known, 2) value returned by CreateDomain
	// (Mailgun generates one when omitted), 3) null - matches what ImportState
	// produces, otherwise ImportStateVerify sees a "" vs null drift.
	switch {
	case !planPwd.IsNull() && !planPwd.IsUnknown():
		plan.SmtpPassword = planPwd
	case createResp.Domain.SMTPPassword != "":
		plan.SmtpPassword = types.StringValue(createResp.Domain.SMTPPassword)
	default:
		plan.SmtpPassword = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	statePwd := state.SmtpPassword
	diags, notFound := refreshDomain(ctx, client, state.ID.ValueString(), &state)
	if notFound {
		log.Printf("[WARN] Mailgun domain %s not found, removing from state", state.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SmtpPassword = statePwd

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state domainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}
	name := plan.Name.ValueString()

	if !plan.SmtpPassword.Equal(state.SmtpPassword) && !plan.SmtpPassword.IsNull() {
		if err := client.ChangeCredentialPassword(ctx, name, state.SmtpLogin.ValueString(), plan.SmtpPassword.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to update SMTP password", err.Error())
			return
		}
	}
	if !plan.OpenTracking.Equal(state.OpenTracking) {
		v := boolToYesNo(plan.OpenTracking.ValueBool())
		if err := client.UpdateOpenTracking(ctx, name, v); err != nil {
			resp.Diagnostics.AddError("Failed to update open tracking", err.Error())
			return
		}
	}
	if !plan.ClickTracking.Equal(state.ClickTracking) {
		v := boolToYesNo(plan.ClickTracking.ValueBool())
		if err := client.UpdateClickTracking(ctx, name, v); err != nil {
			resp.Diagnostics.AddError("Failed to update click tracking", err.Error())
			return
		}
	}
	if !plan.WebScheme.Equal(state.WebScheme) {
		opts := mailgun.UpdateDomainOptions{WebScheme: plan.WebScheme.ValueString()}
		if err := client.UpdateDomain(ctx, name, &opts); err != nil {
			resp.Diagnostics.AddError("Failed to update web scheme", err.Error())
			return
		}
	}

	// Preserve smtp_password from plan (API never returns it).
	planPwd := plan.SmtpPassword
	diags, _ := refreshDomain(ctx, client, name, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.SmtpPassword = planPwd
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.cfg.GetClient(state.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Mailgun client error", err.Error())
		return
	}

	id := state.ID.ValueString()
	log.Printf("[INFO] Deleting Domain: %s", id)
	if err := client.DeleteDomain(ctx, id); err != nil {
		resp.Diagnostics.AddError("Failed to delete domain", err.Error())
		return
	}

	// Poll until the domain disappears (Mailgun is eventually consistent).
	deadline := time.Now().Add(5 * time.Minute)
	for {
		_, err := client.GetDomain(ctx, id, nil)
		if err != nil {
			log.Printf("[INFO] Got error looking for domain, seems gone: %s", err)
			return
		}
		if time.Now().After(deadline) {
			resp.Diagnostics.AddError("Timeout waiting for domain deletion",
				fmt.Sprintf("domain %s still exists after 5 minutes", id))
			return
		}
		log.Printf("[INFO] Retrying until domain disappears...")
		time.Sleep(5 * time.Second)
	}
}

// boolToYesNo converts a bool to the yes/no value expected by Mailgun
// tracking endpoints.
func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
