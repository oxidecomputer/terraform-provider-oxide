// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcinternetgatewayipaddressattachment

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
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
	ID          types.String      `tfsdk:"id"`
	GatewayID   types.String      `tfsdk:"gateway_id"`
	Address     iptypes.IPAddress `tfsdk:"address"`
	Name        types.String      `tfsdk:"name"`
	Description types.String      `tfsdk:"description"`
	Timeouts    timeouts.Value    `tfsdk:"timeouts"`
}

func (r *Resource) Metadata(
	_ context.Context,
	_ resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_vpc_internet_gateway_ip_address_attachment"
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

func (r *Resource) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This resource manages the attachment of an IP address to a VPC internet gateway.",
		Attributes: map[string]schema.Attribute{
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
			"address": schema.StringAttribute{
				Required:    true,
				CustomType:  iptypes.IPAddressType{},
				Description: "IP address to attach to the internet gateway.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the IP address attachment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the IP address attachment.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the IP address attachment.",
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

	address := plan.Address.ValueString()
	attachment, err := r.client.InternetGatewayIpAddressCreate(
		ctx,
		oxide.InternetGatewayIpAddressCreateParams{
			Gateway: oxide.NameOrId(plan.GatewayID.ValueString()),
			Body: &oxide.InternetGatewayIpAddressCreate{
				Address:     address,
				Description: plan.Description.ValueString(),
				Name:        oxide.Name(plan.Name.ValueString()),
			},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching IP address to internet gateway",
			"API error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(attachment.Id)
	plan.Address = iptypes.NewIPAddressValue(attachment.Address)
	plan.Name = types.StringValue(string(attachment.Name))
	plan.Description = types.StringValue(attachment.Description)

	tflog.Trace(
		ctx,
		fmt.Sprintf("attached IP address with attachment ID: %v", attachment.Id),
		map[string]any{"success": true},
	)

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

	attachments, err := r.client.InternetGatewayIpAddressListAllPages(
		ctx,
		oxide.InternetGatewayIpAddressListParams{
			Gateway: oxide.NameOrId(state.GatewayID.ValueString()),
		},
	)
	if err != nil {
		if errors.Is(err, oxide.ErrHTTP404) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read internet gateway IP address attachment",
			"API error: "+err.Error(),
		)
		return
	}

	idx := slices.IndexFunc(attachments, func(attachment oxide.InternetGatewayIpAddress) bool {
		return attachment.Id == state.ID.ValueString()
	})
	if idx == -1 {
		resp.State.RemoveResource(ctx)
		return
	}

	attachment := attachments[idx]
	state.ID = types.StringValue(attachment.Id)
	state.Address = iptypes.NewIPAddressValue(attachment.Address)
	state.Name = types.StringValue(string(attachment.Name))
	state.Description = types.StringValue(attachment.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// All the resource's attributes require replacement. This implementation
	// writes the plan to state so that timeouts can be modified.
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

	err := r.client.InternetGatewayIpAddressDelete(
		ctx,
		oxide.InternetGatewayIpAddressDeleteParams{
			Address: oxide.NameOrId(state.ID.ValueString()),
			Cascade: oxide.NewPointer(false),
		},
	)
	if err != nil && !errors.Is(err, oxide.ErrHTTP404) {
		resp.Diagnostics.AddError(
			"Error detaching IP address from internet gateway",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf(
			"detached internet gateway IP address attachment with ID: %v",
			state.ID.ValueString(),
		),
		map[string]any{"success": true},
	)
}

func parseImportID(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"expected import ID format INTERNET_GATEWAY_ID/ATTACHMENT_ID, got: %s",
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
