// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var (
	_ datasource.DataSource              = (*sshKeyDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*sshKeyDataSource)(nil)
)

// NewSSHKeyDataSource is a helper function to simplify the provider implementation.
func NewSSHKeyDataSource() datasource.DataSource {
	return &sshKeyDataSource{}
}

// sshKeyDataSource is the data source implementation.
type sshKeyDataSource struct {
	client *oxide.Client
}

// sshKeyDataSourceModel are the attributes that are supported on this data source.
type sshKeyDataSourceModel struct {
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	Description  types.String   `tfsdk:"description"`
	PublicKey    types.String   `tfsdk:"public_key"`
	SiloUserID   types.String   `tfsdk:"silo_user_id"`
	TimeCreated  types.String   `tfsdk:"time_created"`
	TimeModified types.String   `tfsdk:"time_modified"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

// Metadata sets the resource type name.
func (d *sshKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_ssh_key"
}

// Configure adds the provider configured client to the data source.
func (d *sshKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

// Schema defines the schema for the data source.
func (d *sshKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the SSH key.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the SSH key.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description for the SSH key.",
			},
			"public_key": schema.StringAttribute{
				Computed:    true,
				Description: "Public SSH key.",
			},
			"silo_user_id": schema.StringAttribute{
				Computed:    true,
				Description: "User ID that owns this SSH key.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this SSH key was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this SSH key was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sshKeyDataSourceModel

	// Read Terraform configuration data into the model.
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
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

	params := oxide.CurrentUserSshKeyViewParams{
		SshKey: oxide.NameOrId(state.Name.ValueString()),
	}
	sshKey, err := d.client.CurrentUserSshKeyView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read SSH key:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read SSH key with ID: %v", sshKey.Id), map[string]any{"success": true})

	state.ID = types.StringValue(sshKey.Id)
	state.Name = types.StringValue(string(sshKey.Name))
	state.Description = types.StringValue(sshKey.Description)
	state.PublicKey = types.StringValue(string(sshKey.PublicKey))
	state.SiloUserID = types.StringValue(string(sshKey.SiloUserId))
	state.TimeCreated = types.StringValue(sshKey.TimeCreated.String())
	state.TimeModified = types.StringValue(sshKey.TimeModified.String())

	// Save retrieved state into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
