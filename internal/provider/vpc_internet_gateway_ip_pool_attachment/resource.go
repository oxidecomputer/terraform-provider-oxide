// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcinternetgatewayippoolattachment

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *oxide.Client
}

type ResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	GatewayID   types.String   `tfsdk:"gateway_id"`
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	IPPoolID    types.String   `tfsdk:"ip_pool_id"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (r *Resource) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpc_internet_gateway_ip_pool_attachment"
}

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

func (r *Resource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	gatewayID, attachmentID, err := parseImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gateway_id"), gatewayID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), attachmentID)...)
}

func parseImportID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"expected import ID format internet_gateway_id/attachment_id, got: %s",
			id,
		)
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		return "", "", fmt.Errorf("internet gateway ID must be a UUID, got: %s", parts[0])
	}
	if _, err := uuid.Parse(parts[1]); err != nil {
		return "", "", fmt.Errorf("attachment ID must be a UUID, got: %s", parts[1])
	}

	return parts[0], parts[1], nil
}

func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource attaches an IP pool to a VPC internet gateway.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the attachment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gateway_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the VPC internet gateway.",
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the IP pool attachment, unique within the internet gateway.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable free-form text about the IP pool attachment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_pool_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the IP pool to attach.",
				Validators: []validator.String{
					oxidevalidator.IsUUID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

	attachment, err := r.client.InternetGatewayIpPoolCreate(
		ctx,
		oxide.InternetGatewayIpPoolCreateParams{
			Gateway: oxide.NameOrId(plan.GatewayID.ValueString()),
			Body: &oxide.InternetGatewayIpPoolCreate{
				Name:        oxide.Name(plan.Name.ValueString()),
				Description: plan.Description.ValueString(),
				IpPool:      oxide.NameOrId(plan.IPPoolID.ValueString()),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Error attaching IP pool", "API error: "+err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("attached IP pool with attachment ID: %v", attachment.Id))

	plan.ID = types.StringValue(attachment.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

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

	attachments, err := r.client.InternetGatewayIpPoolListAllPages(
		ctx,
		oxide.InternetGatewayIpPoolListParams{
			Gateway: oxide.NameOrId(state.GatewayID.ValueString()),
		},
	)
	if err != nil {
		if errors.Is(err, oxide.ErrHTTP404) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to list attached IP pools", "API error: "+err.Error())
		return
	}

	idx := slices.IndexFunc(attachments, func(element oxide.InternetGatewayIpPool) bool {
		return element.Id == state.ID.ValueString()
	})
	if idx == -1 {
		resp.State.RemoveResource(ctx)
		return
	}

	attachment := attachments[idx]

	state.Name = types.StringValue(string(attachment.Name))
	state.Description = types.StringValue(attachment.Description)
	state.IPPoolID = types.StringValue(attachment.IpPoolId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// All the resource's attributes require replace. This update implementation
	// reads the plan and write it to state so that timeouts can be modified.
	var plan ResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

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

	if err := r.client.InternetGatewayIpPoolDelete(
		ctx,
		oxide.InternetGatewayIpPoolDeleteParams{
			Pool:    oxide.NameOrId(state.ID.ValueString()),
			Cascade: oxide.NewPointer(false),
		},
	); err != nil && !errors.Is(err, oxide.ErrHTTP404) {
		resp.Diagnostics.AddError("Error detaching IP pool", "API error: "+err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("detached IP pool attachment with ID: %v", state.ID.ValueString()))
}
