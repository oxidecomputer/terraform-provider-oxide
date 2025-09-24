// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// Compile-time assertions to check that siloResource implements the necessary
// Terraform resource interfaces.
var (
	_ resource.Resource                = (*siloResource)(nil)
	_ resource.ResourceWithConfigure   = (*siloResource)(nil)
	_ resource.ResourceWithImportState = (*siloResource)(nil)
)

// NewSiloResource is a helper to easily construct a siloResource as a type that
// implements the Terraform resource interface.
func NewSiloResource() resource.Resource {
	return &siloResource{}
}

// siloResource is the concrete type that implements the necessary Terraform
// resource interfaces. It holds state to interact with the Oxide API.
type siloResource struct {
	client *oxide.Client
}

// siloResourceModel represents the Terraform configuration and state for the
// Oxide silo resource.
type siloResourceModel struct {
	ID               types.String                      `tfsdk:"id"`
	Name             types.String                      `tfsdk:"name"`
	Description      types.String                      `tfsdk:"description"`
	Quotas           *siloResourceQuotasModel          `tfsdk:"quotas"`
	TlsCertificates  []siloResourceTlsCertificateModel `tfsdk:"tls_certificates"`
	Discoverable     types.Bool                        `tfsdk:"discoverable"`
	IdentityMode     types.String                      `tfsdk:"identity_mode"`
	AdminGroupName   types.String                      `tfsdk:"admin_group_name"`
	MappedFleetRoles map[string][]string               `tfsdk:"mapped_fleet_roles"`
	TimeCreated      types.String                      `tfsdk:"time_created"`
	TimeModified     types.String                      `tfsdk:"time_modified"`
	Timeouts         timeouts.Value                    `tfsdk:"timeouts"`
}

// siloResourceQuotasModel represents quotas for an Oxide silo.
type siloResourceQuotasModel struct {
	Cpus    types.Int64 `tfsdk:"cpus"`
	Memory  types.Int64 `tfsdk:"memory"`
	Storage types.Int64 `tfsdk:"storage"`
}

// siloResourceTlsCertificateModel represents a TLS certificate for an Oxide
// silo.
type siloResourceTlsCertificateModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Cert        types.String `tfsdk:"cert"`
	Key         types.String `tfsdk:"key"`
	Service     types.String `tfsdk:"service"`
}

// Metadata configures the Terraform resource name for this resource.
func (r *siloResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_silo"
}

// Configure sets up necessary data or clients needed by this resource.
func (r *siloResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports this resource using its ID.
func (r *siloResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

const tlsCertificateRegEx = `^[a-zA-Z0-9-]+$`

// Schema defines the attributes for this resource.
func (r *siloResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique, immutable, user-controlled identifier of the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(tlsCertificateRegEx),
						`Names must begin with a lower case ASCII letter, be composed exclusively of lowercase ASCII, uppercase ASCII, numbers, and '-'.`,
					),
					stringvalidator.LengthAtMost(63),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable free-form text about the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"quotas": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Limits the amount of provisionable CPU, memory, and storage in the silo.",
				Attributes: map[string]schema.Attribute{
					"cpus": schema.Int64Attribute{
						Required:    true,
						Description: "Amount of virtual CPUs available for running instances in the silo.",
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
					"memory": schema.Int64Attribute{
						Required:    true,
						Description: "Amount of memory, in bytes, available for running instances in the silo.",
					},
					"storage": schema.Int64Attribute{
						Required:    true,
						Description: "Amount of storage, in bytes, available for disks or snapshots.",
					},
				},
			},
			"tls_certificates": schema.ListNestedAttribute{
				Required:    true,
				WriteOnly:   true,
				Description: "Initial TLS certificates to be used for the new silo's console and API endpoints.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							WriteOnly:   true,
							Description: "Unique, immutable, user-controlled identifier of the certificate.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompile(tlsCertificateRegEx),
									`Names must begin with a lower case ASCII letter, be composed exclusively of lowercase ASCII, uppercase ASCII, numbers, and '-'.`,
								),
								stringvalidator.LengthAtMost(63),
							},
						},
						"description": schema.StringAttribute{
							Required:    true,
							WriteOnly:   true,
							Description: "Human-readable free-form text about the certificate.",
						},
						"cert": schema.StringAttribute{
							Required:    true,
							WriteOnly:   true,
							Description: "PEM-formatted string containing public certificate chain.",
						},
						"key": schema.StringAttribute{
							Required:    true,
							WriteOnly:   true,
							Description: "PEM-formatted string containing private key.",
						},
						"service": schema.StringAttribute{
							Required:    true,
							WriteOnly:   true,
							Description: "Service using this certificate.",
							Validators: []validator.String{
								stringvalidator.OneOf("external_api"),
							},
						},
					},
				},
			},
			"discoverable": schema.BoolAttribute{
				Required:    true,
				Description: "Whether this silo is present in the silo_list output.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"identity_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How users and groups are managed in the silo.",
				Default:     stringdefault.StaticString(string(oxide.SiloIdentityModeLocalOnly)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(oxide.SiloIdentityModeLocalOnly),
						string(oxide.SiloIdentityModeSamlJit),
					),
				},
			},
			"admin_group_name": schema.StringAttribute{
				Optional:    true,
				Description: "If set, this group will be created during Silo creation and granted the 'Silo Admin' role.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mapped_fleet_roles": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Mapped Fleet Roles for the Silo.",
				Default: mapdefault.StaticValue(types.MapValueMust(
					types.ListType{ElemType: types.StringType},
					map[string]attr.Value{},
				)),
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Map{
					mapvalidator.KeysAre(
						stringvalidator.OneOf("admin", "collaborator", "viewer"),
					),
					mapvalidator.ValueListsAre(
						listvalidator.ValueStringsAre(
							stringvalidator.OneOf("admin", "collaborator", "viewer"),
						),
					),
				},
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this silo was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this silo was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

// Create creates this resource using the Oxide API.
func (r *siloResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan siloResourceModel

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

	// The tls_certificates attribute is a write-only attribute which must be retrieved
	// from the configuration instead of the plan. We save it into the plan instead of
	// creating a new variable so that code below can solely use the plan variable.
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("tls_certificates"), &plan.TlsCertificates)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := oxide.SiloCreateParams{
		Body: &oxide.SiloCreate{
			AdminGroupName:   plan.AdminGroupName.ValueString(),
			Description:      plan.Description.ValueString(),
			IdentityMode:     oxide.SiloIdentityMode(plan.IdentityMode.ValueString()),
			Discoverable:     plan.Discoverable.ValueBoolPointer(),
			MappedFleetRoles: stringMapToFleetRoleMap(plan.MappedFleetRoles),
			Name:             oxide.Name(plan.Name.ValueString()),
			Quotas: oxide.SiloQuotasCreate{
				// We can safely dereference all fields within plan.Quotas as they are required fields
				Cpus:    oxide.NewPointer(int(*plan.Quotas.Cpus.ValueInt64Pointer())),
				Memory:  oxide.ByteCount(*plan.Quotas.Memory.ValueInt64Pointer()),
				Storage: oxide.ByteCount(*plan.Quotas.Storage.ValueInt64Pointer()),
			},
			TlsCertificates: tlsCertsModelToCertificateCreateSlice(plan.TlsCertificates),
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

	plan.ID = types.StringValue(silo.Id)
	plan.TimeCreated = types.StringValue(silo.TimeCreated.String())
	plan.TimeModified = types.StringValue(silo.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read fetches this resource using the Oxide API.
func (r *siloResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state siloResourceModel

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
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read Silo:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read silo with ID: %v", silo.Id), map[string]any{"success": true})

	siloQuotas, err := r.client.SiloQuotasView(ctx, oxide.SiloQuotasViewParams{
		Silo: oxide.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read silo quotas:",
			"API error: "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(silo.Id)
	state.Name = types.StringValue(string(silo.Name))
	state.Description = types.StringValue(silo.Description)
	state.Quotas = &siloResourceQuotasModel{
		Cpus:    types.Int64Value(int64(*siloQuotas.Cpus)),
		Memory:  types.Int64Value(int64(siloQuotas.Memory)),
		Storage: types.Int64Value(int64(siloQuotas.Storage)),
	}
	state.Discoverable = types.BoolPointerValue(silo.Discoverable)
	state.IdentityMode = types.StringValue(string(silo.IdentityMode))
	state.MappedFleetRoles = fleetRoleMapToStringMap(silo.MappedFleetRoles)
	state.TimeCreated = types.StringValue(silo.TimeCreated.String())
	state.TimeModified = types.StringValue(silo.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates this resource using the Oxide API. Not all attributes can
// be updated. Refer to [Schema] and the Oxide API documentation for more
// information.
func (r *siloResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan siloResourceModel
	var state siloResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

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
			// We can safely dereference all fields within plan.Quotas as they are required fields
			Cpus:    oxide.NewPointer(int(*plan.Quotas.Cpus.ValueInt64Pointer())),
			Memory:  oxide.ByteCount(*plan.Quotas.Memory.ValueInt64Pointer()),
			Storage: oxide.ByteCount(*plan.Quotas.Storage.ValueInt64Pointer()),
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

	silo, err := r.client.SiloView(ctx, oxide.SiloViewParams{
		Silo: oxide.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating silo quotas",
			"API error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(siloQuotas.SiloId)
	plan.Quotas = &siloResourceQuotasModel{
		Cpus:    types.Int64Value(int64(*siloQuotas.Cpus)),
		Memory:  types.Int64Value(int64(siloQuotas.Memory)),
		Storage: types.Int64Value(int64(siloQuotas.Storage)),
	}
	plan.TimeCreated = types.StringValue(silo.TimeCreated.String())
	plan.TimeModified = types.StringValue(silo.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes this resource using the Oxide API.
func (r *siloResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state siloResourceModel

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

func stringMapToFleetRoleMap(mappedFleetRoles map[string][]string) map[string][]oxide.FleetRole {
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

func tlsCertsModelToCertificateCreateSlice(tlsCertificates []siloResourceTlsCertificateModel) []oxide.CertificateCreate {
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

func fleetRoleMapToStringMap(mappedFleetRoles map[string][]oxide.FleetRole) map[string][]string {
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
