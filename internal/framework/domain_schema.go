package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// domainResourceSchema returns the framework schema for mailgun_domain.
// Defaults and plan modifiers are chosen to produce wire-identical state to
// the legacy SDKv2 schema so users do not see noisy plans after upgrade.
func domainResourceSchema() schema.Schema {
	return schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
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
			"spam_action": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("disabled"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"smtp_login": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"smtp_password": schema.StringAttribute{
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"wildcard": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"dkim_selector": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_dkim_authority": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"open_tracking": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"click_tracking": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"web_scheme": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("http"),
			},
			"dkim_key_size": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"use_automatic_sender_security": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"sending_records_set":   sendingRecordsSetAttribute(),
			"receiving_records_set": receivingRecordsSetAttribute(),
		},
	}
}

func sendingRecordsSetAttribute() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":          schema.StringAttribute{Computed: true},
				"name":        schema.StringAttribute{Computed: true},
				"record_type": schema.StringAttribute{Computed: true},
				"valid":       schema.StringAttribute{Computed: true},
				"value":       schema.StringAttribute{Computed: true},
			},
		},
	}
}

func receivingRecordsSetAttribute() schema.SetNestedAttribute {
	return schema.SetNestedAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.Set{
			setplanmodifier.UseStateForUnknown(),
		},
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":          schema.StringAttribute{Computed: true},
				"priority":    schema.StringAttribute{Computed: true},
				"record_type": schema.StringAttribute{Computed: true},
				"valid":       schema.StringAttribute{Computed: true},
				"value":       schema.StringAttribute{Computed: true},
			},
		},
	}
}
