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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &domainResource{}
	_ resource.ResourceWithConfigure   = &domainResource{}
	_ resource.ResourceWithImportState = &domainResource{}
)

// NewDomainResource is a helper function to simplify the provider implementation.
func NewDomainResource() resource.Resource {
	return &domainResource{}
}

// domainResource is the resource implementation.
type domainResource struct {
	client *Config
}

// domainRecordModel maps domain record data.
type domainRecordModel struct {
	ID         types.String `tfsdk:"id"`
	Priority   types.String `tfsdk:"priority"`
	RecordType types.String `tfsdk:"record_type"`
	Valid      types.String `tfsdk:"valid"`
	Value      types.String `tfsdk:"value"`
	Name       types.String `tfsdk:"name"`
}

// domainResourceModel maps the resource schema data.
type domainResourceModel struct {
	ID                  types.String        `tfsdk:"id"`
	Name                types.String        `tfsdk:"name"`
	Region              types.String        `tfsdk:"region"`
	SpamAction          types.String        `tfsdk:"spam_action"`
	SmtpLogin           types.String        `tfsdk:"smtp_login"`
	SmtpPassword        types.String        `tfsdk:"smtp_password"`
	Wildcard            types.Bool          `tfsdk:"wildcard"`
	DkimSelector        types.String        `tfsdk:"dkim_selector"`
	ForceDkimAuthority  types.Bool          `tfsdk:"force_dkim_authority"`
	OpenTracking        types.Bool          `tfsdk:"open_tracking"`
	ClickTracking       types.Bool          `tfsdk:"click_tracking"`
	WebScheme           types.String        `tfsdk:"web_scheme"`
	ReceivingRecords    []domainRecordModel `tfsdk:"receiving_records"`
	ReceivingRecordsSet []domainRecordModel `tfsdk:"receiving_records_set"`
	SendingRecords      []domainRecordModel `tfsdk:"sending_records"`
	SendingRecordsSet   []domainRecordModel `tfsdk:"sending_records_set"`
	DkimKeySize         types.Int64         `tfsdk:"dkim_key_size"`
}

// Metadata returns the resource type name.
func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the resource.
func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Mailgun domain resource. This can be used to create and manage domains on Mailgun.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the domain (same as name).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The domain to add to Mailgun.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("us"),
				Description: "The region where domain will be created. Valid values are 'us' or 'eu'. Default is 'us'.",
				Validators: []validator.String{
					stringvalidator.OneOf("us", "eu"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"spam_action": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("disabled"),
				Description: "The spam action to be used. Valid values are 'disabled', 'block', or 'tag'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"smtp_login": schema.StringAttribute{
				Computed:    true,
				Description: "The SMTP login for the domain.",
			},
			"smtp_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The password for SMTP login.",
			},
			"wildcard": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether the domain is a wildcard domain.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"dkim_selector": schema.StringAttribute{
				Optional:    true,
				Description: "The DKIM selector for the domain.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_dkim_authority": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to force DKIM authority.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"open_tracking": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to enable open tracking.",
			},
			"click_tracking": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to enable click tracking.",
			},
			"web_scheme": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("http"),
				Description: "The web scheme for the domain. Valid values are 'http' or 'https'.",
			},
			"dkim_key_size": schema.Int64Attribute{
				Optional:    true,
				Description: "The size of the DKIM key.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"receiving_records": schema.ListNestedAttribute{
				Computed:           true,
				DeprecationMessage: "Use `receiving_records_set` instead.",
				Description:        "The receiving DNS records for the domain.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the record.",
						},
						"priority": schema.StringAttribute{
							Computed:    true,
							Description: "The priority of the record.",
						},
						"record_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the record.",
						},
						"valid": schema.StringAttribute{
							Computed:    true,
							Description: "Whether the record is valid.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the record.",
						},
					},
				},
			},
			"receiving_records_set": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The receiving DNS records for the domain as a set.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the record.",
						},
						"priority": schema.StringAttribute{
							Computed:    true,
							Description: "The priority of the record.",
						},
						"record_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the record.",
						},
						"valid": schema.StringAttribute{
							Computed:    true,
							Description: "Whether the record is valid.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the record.",
						},
					},
				},
			},
			"sending_records": schema.ListNestedAttribute{
				Computed:           true,
				DeprecationMessage: "Use `sending_records_set` instead.",
				Description:        "The sending DNS records for the domain.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the record.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the record.",
						},
						"record_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the record.",
						},
						"valid": schema.StringAttribute{
							Computed:    true,
							Description: "Whether the record is valid.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the record.",
						},
					},
				},
			},
			"sending_records_set": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The sending DNS records for the domain as a set.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the record.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the record.",
						},
						"record_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the record.",
						},
						"valid": schema.StringAttribute{
							Computed:    true,
							Description: "Whether the record is valid.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the record.",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainResourceModel
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

	opts := mailgun.CreateDomainOptions{}
	opts.SpamAction = mtypes.SpamAction(plan.SpamAction.ValueString())
	if !plan.SmtpPassword.IsNull() {
		opts.Password = plan.SmtpPassword.ValueString()
	}
	if !plan.Wildcard.IsNull() {
		opts.Wildcard = plan.Wildcard.ValueBool()
	}
	if !plan.DkimKeySize.IsNull() {
		opts.DKIMKeySize = int(plan.DkimKeySize.ValueInt64())
	}
	if !plan.ForceDkimAuthority.IsNull() {
		opts.ForceDKIMAuthority = plan.ForceDkimAuthority.ValueBool()
	}
	opts.WebScheme = plan.WebScheme.ValueString()

	log.Printf("[DEBUG] Domain create configuration: %#v", opts)

	_, err = client.CreateDomain(ctx, plan.Name.ValueString(), &opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating domain",
			"Could not create domain: "+err.Error(),
		)
		return
	}

	if !plan.DkimSelector.IsNull() && plan.DkimSelector.ValueString() != "" {
		err = client.UpdateDomainDkimSelector(ctx, plan.Name.ValueString(), plan.DkimSelector.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating DKIM selector",
				"Could not update DKIM selector: "+err.Error(),
			)
			return
		}
	}

	if plan.OpenTracking.ValueBool() {
		err = client.UpdateOpenTracking(ctx, plan.Name.ValueString(), "yes")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating open tracking",
				"Could not update open tracking: "+err.Error(),
			)
			return
		}
	}

	if plan.ClickTracking.ValueBool() {
		err = client.UpdateClickTracking(ctx, plan.Name.ValueString(), "yes")
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating click tracking",
				"Could not update click tracking: "+err.Error(),
			)
			return
		}
	}

	plan.ID = plan.Name
	log.Printf("[INFO] Domain ID: %s", plan.ID.ValueString())

	// Retrieve and update state of domain
	err = r.readDomain(ctx, plan.ID.ValueString(), client, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domain",
			"Could not read domain after creation: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainResourceModel
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

	err = r.readDomain(ctx, state.ID.ValueString(), client, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domain",
			"Could not read domain: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state domainResourceModel
	diags = req.State.Get(ctx, &state)
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

	name := plan.Name.ValueString()

	// Update SMTP password if changed
	if !plan.SmtpPassword.Equal(state.SmtpPassword) && !plan.SmtpPassword.IsNull() {
		err = client.ChangeCredentialPassword(ctx, name, plan.SmtpLogin.ValueString(), plan.SmtpPassword.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating SMTP password",
				"Could not update SMTP password: "+err.Error(),
			)
			return
		}
	}

	// Update open tracking if changed
	if !plan.OpenTracking.Equal(state.OpenTracking) {
		value := "no"
		if plan.OpenTracking.ValueBool() {
			value = "yes"
		}
		err = client.UpdateOpenTracking(ctx, name, value)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating open tracking",
				"Could not update open tracking: "+err.Error(),
			)
			return
		}
	}

	// Update click tracking if changed
	if !plan.ClickTracking.Equal(state.ClickTracking) {
		value := "no"
		if plan.ClickTracking.ValueBool() {
			value = "yes"
		}
		err = client.UpdateClickTracking(ctx, name, value)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating click tracking",
				"Could not update click tracking: "+err.Error(),
			)
			return
		}
	}

	// Update web scheme if changed
	if !plan.WebScheme.Equal(state.WebScheme) {
		opts := mailgun.UpdateDomainOptions{
			WebScheme: plan.WebScheme.ValueString(),
		}
		err = client.UpdateDomain(ctx, name, &opts)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating web scheme",
				"Could not update web scheme: "+err.Error(),
			)
			return
		}
	}

	// Read the updated domain
	err = r.readDomain(ctx, plan.ID.ValueString(), client, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domain",
			"Could not read domain after update: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainResourceModel
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

	log.Printf("[INFO] Deleting Domain: %s", state.ID.ValueString())

	err = client.DeleteDomain(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting domain",
			"Could not delete domain: "+err.Error(),
		)
		return
	}

	// Wait for domain to be deleted
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		_, err = client.GetDomain(ctx, state.ID.ValueString(), nil)
		if err != nil {
			// Domain is gone
			log.Printf("[INFO] Domain deleted successfully: %s", state.ID.ValueString())
			return
		}
		log.Printf("[INFO] Waiting for domain to be deleted...")
		time.Sleep(5 * time.Second)
	}

	resp.Diagnostics.AddWarning(
		"Domain deletion timeout",
		"Domain may still exist after timeout",
	)
}

// ImportState imports the resource into Terraform state.
func (r *domainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)

	var region, domainName string
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		region = parts[0]
		domainName = parts[1]
	} else {
		region = "us"
		domainName = req.ID
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("region"), region)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), domainName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), domainName)...)
}

// readDomain reads the domain data from Mailgun API and populates the model.
func (r *domainResource) readDomain(ctx context.Context, id string, client *mailgun.Client, model *domainResourceModel) error {
	domainResp, err := client.GetDomain(ctx, id, nil)
	if err != nil {
		return fmt.Errorf("error retrieving domain: %w", err)
	}

	model.Name = types.StringValue(domainResp.Domain.Name)
	model.SmtpLogin = types.StringValue(domainResp.Domain.SMTPLogin)
	model.Wildcard = types.BoolValue(domainResp.Domain.Wildcard)
	model.SpamAction = types.StringValue(string(domainResp.Domain.SpamAction))
	model.WebScheme = types.StringValue(domainResp.Domain.WebScheme)

	// Process receiving records
	receivingRecords := make([]domainRecordModel, len(domainResp.ReceivingDNSRecords))
	for i, r := range domainResp.ReceivingDNSRecords {
		receivingRecords[i] = domainRecordModel{
			ID:         types.StringValue(r.Value),
			Priority:   types.StringValue(r.Priority),
			Valid:      types.StringValue(r.Valid),
			Value:      types.StringValue(r.Value),
			RecordType: types.StringValue(r.RecordType),
		}
	}
	model.ReceivingRecords = receivingRecords
	model.ReceivingRecordsSet = receivingRecords

	// Process sending records
	sendingRecords := make([]domainRecordModel, len(domainResp.SendingDNSRecords))
	for i, r := range domainResp.SendingDNSRecords {
		id := r.Name
		if strings.Contains(r.Name, "._domainkey.") {
			id = "_domainkey." + domainResp.Domain.Name
		}
		sendingRecords[i] = domainRecordModel{
			ID:         types.StringValue(id),
			Name:       types.StringValue(r.Name),
			Valid:      types.StringValue(r.Valid),
			Value:      types.StringValue(r.Value),
			RecordType: types.StringValue(r.RecordType),
		}
	}
	model.SendingRecords = sendingRecords
	model.SendingRecordsSet = sendingRecords

	// Get tracking info
	info, err := client.GetDomainTracking(ctx, id)
	if err == nil {
		model.OpenTracking = types.BoolValue(info.Open.Active)
		model.ClickTracking = types.BoolValue(info.Click.Active)
	}

	return nil
}
