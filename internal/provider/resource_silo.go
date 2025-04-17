// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*siloResource)(nil)
	_ resource.ResourceWithConfigure = (*siloResource)(nil)
)

// NewSiloResource is a helper function to simplify the provider implementation.
func NewSiloResource() resource.Resource {
	return &siloResource{}
}

// siloResource is the resource implementation.
type siloResource struct {
	client *oxide.Client
}

type siloResourceModel struct {
	AdminGroupName   types.String                      `tfsdk:"admin_group_name"`
	Description      types.String                      `tfsdk:"description"`
	Discoverable     types.Bool                        `tfsdk:"discoverable"`
	ID               types.String                      `tfsdk:"id"`
	IdentityMode     types.String                      `tfsdk:"identity_mode"`
	MappedFleetRoles map[string][]string               `tfsdk:"mapped_fleet_roles"`
	Name             types.String                      `tfsdk:"name"`
	Quotas           siloResourceQuotaModel            `tfsdk:"quotas"`
	TlsCertificates  []siloResourceTlsCertificateModel `tfsdk:"tls_certificates"`
	Timeouts         timeouts.Value                    `tfsdk:"timeouts"`
	TimeCreated      types.String                      `tfsdk:"time_created"`
	TimeModified     types.String                      `tfsdk:"time_modified"`
}

type siloResourceQuotaModel struct {
	Cpus    types.Int64 `tfsdk:"cpus"`
	Memory  types.Int64 `tfsdk:"memory"`
	Storage types.Int64 `tfsdk:"storage"`
}

type siloResourceTlsCertificateModel struct {
	Cert        types.String `tfsdk:"cert"`
	Description types.String `tfsdk:"description"`
	Key         types.String `tfsdk:"key"`
	Name        types.String `tfsdk:"name"`
	Service     types.String `tfsdk:"service"`
}

// Metadata returns the resource type name.
func (r *siloResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_silo"
}

// Configure adds the provider configured client to the data source.
func (r *siloResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *siloResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Schema defines the schema for the resource.
func (r *siloResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"admin_group_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the group to be granted the Silo Admin role.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"discoverable": schema.BoolAttribute{
				Required:    true,
				Description: "Whether this silo is present in the silo_list output.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"mapped_fleet_roles": schema.MapAttribute{
				Optional:    true,
				Description: "Mapped Fleet Roles for the Silo.",
				ElementType: types.ListType{ElemType: types.StringType},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"quotas": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Limits the amount of provisionable CPU, memory, and storage in the Silo.",
				Attributes: map[string]schema.Attribute{
					"cpus": schema.Int64Attribute{
						Required:    true,
						Description: "Amount of virtual CPUs available for running instances in the Silo.",
					},
					"memory": schema.Int64Attribute{
						Required:    true,
						Description: "Amount of RAM (in bytes) available for running instances in the Silo.",
					},
					"storage": schema.Int64Attribute{
						Required:    true,
						Description: "Amount of storage (in bytes) available for disks or snapshots.",
					},
				},
			},
			"tls_certificates": schema.ListNestedAttribute{
				Required:    true,
				Description: "Initial TLS certificates to be used for the new Silo's console and API endpoints.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cert": schema.StringAttribute{
							Description: "PEM-formatted string containing public certificate chain.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"description": schema.StringAttribute{
							Description: "Certificate description.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"key": schema.StringAttribute{
							Description: "PEM-formatted string containing private key.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"name": schema.StringAttribute{
							Description: "Name associated to the certificate.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9-]+$`),
									"allowed characters are lowercase ASCII, digits, and \"-\""),
							},
						},
						"service": schema.StringAttribute{
							Description: "Service using this certificate.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf("external_api"),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this silo was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this silo was last modified.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the silo.",
			},
			"identity_mode": schema.StringAttribute{
				Description: "How users and groups are managed in the Silo",
				Required:    true,
				Default:     stringdefault.StaticString(string(oxide.SiloIdentityModeLocalOnly)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(oxide.SiloIdentityModeLocalOnly),
						string(oxide.SiloIdentityModeSamlJit),
					),
				},
			},
		},
	}
}

func toOxideMappedFleetRoles(mappedFleetRoles map[string][]string) map[string][]oxide.FleetRole {
	model := make(map[string][]oxide.FleetRole)

	for key, fleetRoleModels := range mappedFleetRoles {
		var roles []oxide.FleetRole
		for _, frm := range fleetRoleModels {
			roles = append(roles, oxide.FleetRole(frm))
		}
		model[key] = roles
	}
	return model
}

func toOxideTlsCertificates(tlsCertificates []siloResourceTlsCertificateModel) []oxide.CertificateCreate {
	var model []oxide.CertificateCreate

	for _, tlsCert := range tlsCertificates {
		r := oxide.CertificateCreate{
			Cert:        tlsCert.Cert.ValueString(),
			Description: tlsCert.Description.ValueString(),
			Key:         tlsCert.Key.ValueString(),
			Name:        oxide.Name(tlsCert.Name.ValueString()),
			Service:     oxide.ServiceUsingCertificate(tlsCert.Service.ValueString()),
		}

		model = append(model, r)
	}

	return model
}

// Create creates the resource and sets the initial Terraform state.
func (r *siloResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan siloResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	params := oxide.SiloCreateParams{
		Body: &oxide.SiloCreate{
			AdminGroupName:   plan.AdminGroupName.ValueString(),
			Description:      plan.Description.ValueString(),
			IdentityMode:     oxide.SiloIdentityMode(plan.IdentityMode.ValueString()),
			Discoverable:     plan.Discoverable.ValueBoolPointer(),
			MappedFleetRoles: toOxideMappedFleetRoles(plan.MappedFleetRoles),
			Name:             oxide.Name(plan.Name.ValueString()),
			Quotas: oxide.SiloQuotasCreate{
				Cpus:    int(plan.Quotas.Cpus.ValueInt64()),
				Memory:  oxide.ByteCount(plan.Quotas.Memory.ValueInt64()),
				Storage: oxide.ByteCount(plan.Quotas.Storage.ValueInt64()),
			},
			TlsCertificates: toOxideTlsCertificates(plan.TlsCertificates),
		},
	}
	tflog.Debug(ctx, fmt.Sprintf("Silo creation parameters: %+v", params.Body.TlsCertificates), nil)
	silo, err := r.client.SiloCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating silo",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created silo with ID: %v", silo.Id), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(silo.Id)
	plan.TimeCreated = types.StringValue(silo.TimeCreated.String())
	plan.TimeModified = types.StringValue(silo.TimeModified.String())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *siloResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state siloResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	params := oxide.SiloViewParams{
		Silo: oxide.NameOrId(state.ID.ValueString()),
	}
	silo, err := r.client.SiloView(ctx, params)
	if err != nil {
		if is404(err) {
			// Remove resource from state during a refresh
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read Silo:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read Silo with ID: %v", silo.Id), map[string]any{"success": true})

	state.Description = types.StringValue(silo.Description)
	state.Discoverable = types.BoolPointerValue(silo.Discoverable)
	state.ID = types.StringValue(silo.Id)
	state.IdentityMode = types.StringValue(string(silo.IdentityMode))
	state.MappedFleetRoles = toTerraformMappedFleetRoles(silo.MappedFleetRoles)
	state.Name = types.StringValue(string(silo.Name))
	state.TimeCreated = types.StringValue(silo.TimeCreated.String())
	state.TimeModified = types.StringValue(silo.TimeModified.String())

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func toTerraformMappedFleetRoles(mappedFleetRoles map[string][]oxide.FleetRole) map[string][]string {
	model := make(map[string][]string)
	for key, roles := range mappedFleetRoles {
		var modelRoles []string
		for _, role := range roles {
			modelRoles = append(modelRoles, string(role))
		}
		model[key] = modelRoles
	}
	return model
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *siloResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan siloResourceModel
	var state siloResourceModel

	// Read Terraform plan data into the plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform prior state data into the state model to retrieve ID
	// which is a computed attribute, so it won't show up in the plan.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	siloQuotasParams := oxide.SiloQuotasUpdateParams{
		Silo: oxide.NameOrId(state.ID.ValueString()),
		Body: &oxide.SiloQuotasUpdate{
			Cpus:    int(plan.Quotas.Cpus.ValueInt64()),
			Memory:  oxide.ByteCount(plan.Quotas.Memory.ValueInt64()),
			Storage: oxide.ByteCount(plan.Quotas.Storage.ValueInt64()),
		},
	}
	siloQuotas, err := r.client.SiloQuotasUpdate(ctx, siloQuotasParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating silo quotas",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated silo with ID: %v", siloQuotas.SiloId), map[string]any{"success": true})

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(siloQuotas.SiloId)
	plan.Quotas.Cpus = types.Int64Value(plan.Quotas.Cpus.ValueInt64())
	plan.Quotas.Memory = types.Int64Value(plan.Quotas.Memory.ValueInt64())
	plan.Quotas.Storage = types.Int64Value(plan.Quotas.Storage.ValueInt64())

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *siloResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state siloResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	params := oxide.SiloDeleteParams{
		Silo: oxide.NameOrId(state.ID.ValueString()),
	}
	if err := r.client.SiloDelete(ctx, params); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Error deleting silo:",
				"API error: "+err.Error(),
			)
			return
		}
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted silo with ID: %v", state.ID.ValueString()), map[string]any{"success": true})
}
