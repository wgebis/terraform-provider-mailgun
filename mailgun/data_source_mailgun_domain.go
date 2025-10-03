package mailgun

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mailgun/mailgun-go/v5"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &domainDataSource{}
	_ datasource.DataSourceWithConfigure = &domainDataSource{}
)

// NewDomainDataSource is a helper function to simplify the provider implementation.
func NewDomainDataSource() datasource.DataSource {
	return &domainDataSource{}
}

// domainDataSource is the data source implementation.
type domainDataSource struct {
	client *Config
}

// domainDataSourceModel maps the data source schema data.
type domainDataSourceModel struct {
	ID                  types.String            `tfsdk:"id"`
	Name                types.String            `tfsdk:"name"`
	Region              types.String            `tfsdk:"region"`
	SpamAction          types.String            `tfsdk:"spam_action"`
	SmtpLogin           types.String            `tfsdk:"smtp_login"`
	SmtpPassword        types.String            `tfsdk:"smtp_password"`
	Wildcard            types.Bool              `tfsdk:"wildcard"`
	DkimSelector        types.String            `tfsdk:"dkim_selector"`
	ForceDkimAuthority  types.Bool              `tfsdk:"force_dkim_authority"`
	OpenTracking        types.Bool              `tfsdk:"open_tracking"`
	ClickTracking       types.Bool              `tfsdk:"click_tracking"`
	WebScheme           types.String            `tfsdk:"web_scheme"`
	ReceivingRecords    []domainRecordDataModel `tfsdk:"receiving_records"`
	ReceivingRecordsSet []domainRecordDataModel `tfsdk:"receiving_records_set"`
	SendingRecords      []domainRecordDataModel `tfsdk:"sending_records"`
	SendingRecordsSet   []domainRecordDataModel `tfsdk:"sending_records_set"`
	DkimKeySize         types.Int64             `tfsdk:"dkim_key_size"`
}

// domainRecordDataModel maps domain record data for data sources.
type domainRecordDataModel struct {
	ID         types.String `tfsdk:"id"`
	Priority   types.String `tfsdk:"priority"`
	RecordType types.String `tfsdk:"record_type"`
	Valid      types.String `tfsdk:"valid"`
	Value      types.String `tfsdk:"value"`
	Name       types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the data source.
func (d *domainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get details about a Mailgun domain.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the domain (same as name).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The domain name to retrieve.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "The region where domain is located. Valid values are 'us' or 'eu'. Default is 'us'.",
				Validators: []validator.String{
					stringvalidator.OneOf("us", "eu"),
				},
			},
			"spam_action": schema.StringAttribute{
				Computed:    true,
				Description: "The spam action configured for the domain.",
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
				Computed:    true,
				Description: "Whether the domain is a wildcard domain.",
			},
			"dkim_selector": schema.StringAttribute{
				Optional:    true,
				Description: "The DKIM selector for the domain.",
			},
			"force_dkim_authority": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to force DKIM authority.",
			},
			"open_tracking": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether open tracking is enabled.",
			},
			"click_tracking": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether click tracking is enabled.",
			},
			"web_scheme": schema.StringAttribute{
				Computed:    true,
				Description: "The web scheme for the domain.",
			},
			"dkim_key_size": schema.Int64Attribute{
				Optional:    true,
				Description: "The size of the DKIM key.",
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
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the record.",
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
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the record.",
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
						"priority": schema.StringAttribute{
							Computed:    true,
							Description: "The priority of the record.",
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
						"priority": schema.StringAttribute{
							Computed:    true,
							Description: "The priority of the record.",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read retrieves data from the Mailgun API and sets the Terraform state.
func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config domainDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default region to "us" if not specified
	region := "us"
	if !config.Region.IsNull() {
		region = config.Region.ValueString()
	}

	client, err := d.client.GetClient(region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Mailgun client",
			"Could not create Mailgun client: "+err.Error(),
		)
		return
	}

	name := config.Name.ValueString()

	err = d.readDomain(ctx, name, client, &config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading domain",
			"Could not read domain: "+err.Error(),
		)
		return
	}

	config.ID = types.StringValue(name)
	config.Region = types.StringValue(region)

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

// readDomain reads the domain data from Mailgun API and populates the model.
func (d *domainDataSource) readDomain(ctx context.Context, name string, client *mailgun.Client, model *domainDataSourceModel) error {
	domainResp, err := client.GetDomain(ctx, name, nil)
	if err != nil {
		return fmt.Errorf("error retrieving domain: %w", err)
	}

	model.Name = types.StringValue(domainResp.Domain.Name)
	model.SmtpLogin = types.StringValue(domainResp.Domain.SMTPLogin)
	model.Wildcard = types.BoolValue(domainResp.Domain.Wildcard)
	model.SpamAction = types.StringValue(string(domainResp.Domain.SpamAction))
	model.WebScheme = types.StringValue(domainResp.Domain.WebScheme)

	// Process receiving records
	receivingRecords := make([]domainRecordDataModel, len(domainResp.ReceivingDNSRecords))
	for i, r := range domainResp.ReceivingDNSRecords {
		receivingRecords[i] = domainRecordDataModel{
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
	sendingRecords := make([]domainRecordDataModel, len(domainResp.SendingDNSRecords))
	for i, r := range domainResp.SendingDNSRecords {
		id := r.Name
		if len(r.Name) > 0 && len(domainResp.Domain.Name) > 0 {
			// Handle DKIM records specially
			if r.RecordType == "TXT" && len(r.Name) > len(domainResp.Domain.Name) {
				id = "_domainkey." + domainResp.Domain.Name
			}
		}
		sendingRecords[i] = domainRecordDataModel{
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
	info, err := client.GetDomainTracking(ctx, name)
	if err == nil {
		model.OpenTracking = types.BoolValue(info.Open.Active)
		model.ClickTracking = types.BoolValue(info.Click.Active)
	}

	return nil
}
