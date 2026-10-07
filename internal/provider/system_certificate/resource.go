// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemcertificate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
	oxidevalidator "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/validator"
)

var (
	_ resource.Resource                = (*Resource)(nil)
	_ resource.ResourceWithConfigure   = (*Resource)(nil)
	_ resource.ResourceWithImportState = (*Resource)(nil)
)

// NewResource returns the system certificate resource.
func NewResource() resource.Resource {
	return &Resource{}
}

// Resource manages a TLS certificate for an Oxide silo.
type Resource struct {
	client *oxide.Client
}

// ResourceModel represents the Terraform configuration and state for the
// system certificate resource.
type ResourceModel struct {
	ID           types.String   `tfsdk:"id"`
	SiloID       types.String   `tfsdk:"silo_id"`
	Name         types.String   `tfsdk:"name"`
	Description  types.String   `tfsdk:"description"`
	Cert         types.String   `tfsdk:"cert"`
	Key          types.String   `tfsdk:"key"`
	Service      types.String   `tfsdk:"service"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// Metadata sets the resource type name.
func (r *Resource) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_system_certificate"
}

// Configure adds the provider configured client to the resource.
func (r *Resource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

// ImportState imports this resource using a composite ID in the format
// `${SILO}/${CERTIFICATE}`, where each part is a name or ID. The certificate
// API does not return the silo, so the certificate is looked up within the silo
// to verify that it belongs to the silo.
func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	silo, certificate, ok := strings.Cut(req.ID, "/")
	if !ok || silo == "" || certificate == "" ||
		strings.Contains(certificate, "/") {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf(
				"Expected import ID format: silo/certificate, got: %s",
				req.ID,
			),
		)
		return
	}

	s, err := r.client.SiloView(
		ctx,
		oxide.SiloViewParams{Silo: oxide.NameOrId(silo)},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read silo:",
			"API error: "+err.Error(),
		)
		return
	}

	// The silo must be omitted when the certificate is selected by ID, so resolve
	// the certificate's name to look it up within the silo below.
	name, certificateID := certificate, ""
	if _, err := uuid.Parse(certificate); err == nil {
		cert, err := r.client.SiloCertificateView(
			ctx,
			oxide.SiloCertificateViewParams{
				Certificate: oxide.NameOrId(certificate),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to read silo certificate:",
				"API error: "+err.Error(),
			)
			return
		}
		name, certificateID = string(cert.Name), cert.Id
	}

	cert, err := r.client.SiloCertificateView(
		ctx,
		oxide.SiloCertificateViewParams{
			Silo:        oxide.NameOrId(s.Id),
			Certificate: oxide.NameOrId(name),
		},
	)
	if err != nil && !errors.Is(err, oxide.ErrHTTP404) {
		resp.Diagnostics.AddError(
			"Unable to read silo certificate:",
			"API error: "+err.Error(),
		)
		return
	}
	if err != nil || (certificateID != "" && cert.Id != certificateID) {
		resp.Diagnostics.AddError(
			"Silo Certificate Not Found",
			fmt.Sprintf(
				"Certificate %q does not belong to silo %q.",
				certificate,
				silo,
			),
		)
		return
	}

	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("silo_id"), s.Id)...,
	)
	resp.Diagnostics.Append(
		resp.State.SetAttribute(ctx, path.Root("id"), cert.Id)...,
	)
}

// Schema defines the attributes for this resource.
func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: shared.ReplaceBackticks(`
This resource manages a single TLS certificate for an Oxide silo, which is used
for the silo's console and API endpoints.

Creating a certificate using this resource does not affect other certificates
for the silo, such as those created with the ''tls_certificates'' attribute of
the ''oxide_system_silo'' resource.

-> Certificates cannot be updated. Changes to the attributes of this resource
will result in the certificate being destroyed and created anew.

~> To rotate a certificate, first add the new certificate as a separate
''oxide_system_certificate'' resource with a different name and apply it. Then
remove the resource for the existing certificate. Replacing a certificate in
place destroys it before its replacement is created, which can leave the silo
without a valid certificate if the replacement is rejected. Consider using the
''prevent_destroy'' or ''create_before_destroy'' lifecycle attributes to have
more control over the resource lifecycle.
`),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"silo_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the silo that uses the certificate.",
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique, immutable, user-controlled identifier of the certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable free-form text about the certificate.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cert": schema.StringAttribute{
				Required:    true,
				Description: "PEM-formatted string containing public certificate chain.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Required:  true,
				WriteOnly: true,
				Sensitive: true,
				MarkdownDescription: "PEM-formatted string containing private key. " +
					"Changes to this attribute alone are not detected, " +
					"as a new private key requires a new `cert`.",
			},
			"service": schema.StringAttribute{
				Required:    true,
				Description: "Service using this certificate. Currently, only `external_api` is supported.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(oxide.ServiceUsingCertificateExternalApi),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this certificate was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this certificate was last modified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Delete: true,
			}),
		},
	}
}

// Create creates this resource using the Oxide API.
func (r *Resource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// The key attribute is write-only and must be retrieved from the
	// configuration instead of the plan.
	var key types.String
	resp.Diagnostics.Append(
		req.Config.GetAttribute(ctx, path.Root("key"), &key)...,
	)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := r.client.SiloCertificateCreate(
		ctx,
		oxide.SiloCertificateCreateParams{
			Silo: oxide.NameOrId(plan.SiloID.ValueString()),
			Body: &oxide.CertificateCreate{
				Name:        oxide.Name(plan.Name.ValueString()),
				Description: plan.Description.ValueString(),
				Cert:        plan.Cert.ValueString(),
				Key:         key.ValueString(),
				Service: oxide.ServiceUsingCertificate(
					plan.Service.ValueString(),
				),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating silo certificate",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("created silo certificate with ID: %v", cert.Id),
		map[string]any{"success": true},
	)

	plan.ID = types.StringValue(cert.Id)
	plan.TimeCreated = types.StringValue(cert.TimeCreated.String())
	plan.TimeModified = types.StringValue(cert.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read fetches this resource using the Oxide API.
func (r *Resource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	// The silo must be omitted when the certificate is selected by ID.
	cert, err := r.client.SiloCertificateView(
		ctx,
		oxide.SiloCertificateViewParams{
			Certificate: oxide.NameOrId(state.ID.ValueString()),
		},
	)
	if err != nil {
		if errors.Is(err, oxide.ErrHTTP404) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read silo certificate:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read silo certificate with ID: %v", cert.Id),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(cert.Id)
	state.Name = types.StringValue(string(cert.Name))
	state.Description = types.StringValue(cert.Description)
	state.Cert = types.StringValue(cert.Cert)
	state.Service = types.StringValue(string(cert.Service))
	state.TimeCreated = types.StringValue(cert.TimeCreated.String())
	state.TimeModified = types.StringValue(cert.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update only saves the planned state since certificates do not have an
// update API. All configurable attributes other than the write-only key
// require replacement, so only changes to timeouts reach this method.
func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes this resource using the Oxide API.
func (r *Resource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, shared.DefaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// The silo must be omitted when the certificate is selected by ID.
	if err := r.client.SiloCertificateDelete(
		ctx,
		oxide.SiloCertificateDeleteParams{
			Certificate: oxide.NameOrId(state.ID.ValueString()),
		},
	); err != nil {
		if !errors.Is(err, oxide.ErrHTTP404) {
			resp.Diagnostics.AddError(
				"Error deleting silo certificate:",
				"API error: "+err.Error(),
			)
			return
		}
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("deleted silo certificate with ID: %v", state.ID.ValueString()),
		map[string]any{"success": true},
	)
}
