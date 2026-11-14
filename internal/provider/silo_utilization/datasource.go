// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package siloutilization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

var _ datasource.DataSource = (*DataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

type DataSource struct {
	client *oxide.Client
}

type DataSourceModel struct {
	Capacity    *ResourceCountsModel `tfsdk:"capacity"`
	Provisioned *ResourceCountsModel `tfsdk:"provisioned"`
	Timeouts    timeouts.Value       `tfsdk:"timeouts"`
}

type ResourceCountsModel struct {
	Cpus    types.Int64 `tfsdk:"cpus"`
	Memory  types.Int64 `tfsdk:"memory"`
	Storage types.Int64 `tfsdk:"storage"`
}

// NewDataSource constructs the data source.
func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// Metadata returns the data source type name.
func (d *DataSource) Metadata(
	_ context.Context,
	_ datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_silo_utilization"
}

// Configure adds the provider configured client to the data source.
func (d *DataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

// Schema defines the schema for the data source.
func (d *DataSource) Schema(
	ctx context.Context,
	_ datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resourceCountsAttributes := map[string]schema.Attribute{
		"cpus": schema.Int64Attribute{
			Computed:    true,
			Description: "Number of virtual CPUs.",
		},
		"memory": schema.Int64Attribute{
			Computed:    true,
			Description: "Amount of memory in bytes.",
		},
		"storage": schema.Int64Attribute{
			Computed:    true,
			Description: "Amount of disk storage in bytes.",
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch resource utilization for the current silo.",
		Attributes: map[string]schema.Attribute{
			"capacity": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Total resources that can be provisioned in the silo.",
				Attributes:  resourceCountsAttributes,
			},
			"provisioned": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Resources allocated to running instances, disks, or snapshots.",
				Attributes:  resourceCountsAttributes,
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *DataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state DataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
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

	utilization, err := d.client.UtilizationView(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read silo utilization:",
			err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "read silo utilization", map[string]any{"success": true})

	state.Capacity = resourceCountsModel(utilization.Capacity)
	state.Provisioned = resourceCountsModel(utilization.Provisioned)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func resourceCountsModel(counts oxide.VirtualResourceCounts) *ResourceCountsModel {
	return &ResourceCountsModel{
		Cpus:    types.Int64Value(int64(*counts.Cpus)),
		Memory:  types.Int64Value(int64(counts.Memory)),
		Storage: types.Int64Value(int64(counts.Storage)),
	}
}
