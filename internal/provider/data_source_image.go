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
	_ datasource.DataSource              = (*imageDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*imageDataSource)(nil)
)

// NewImageDataSource initialises an images datasource
func NewImageDataSource() datasource.DataSource {
	return &imageDataSource{}
}

type imageDataSource struct {
	client *oxide.Client
}

type imageDataSourceModel struct {
	ID           types.String                `tfsdk:"id"`
	ProjectName  types.String                `tfsdk:"project_name"`
	ProjectID    types.String                `tfsdk:"project_id"`
	Timeouts     timeouts.Value              `tfsdk:"timeouts"`
	BlockSize    types.Int64                 `tfsdk:"block_size"`
	Description  types.String                `tfsdk:"description"`
	Digest       *imageDataSourceDigestModel `tfsdk:"digest"`
	Name         types.String                `tfsdk:"name"`
	OS           types.String                `tfsdk:"os"`
	Size         types.Int64                 `tfsdk:"size"`
	TimeCreated  types.String                `tfsdk:"time_created"`
	TimeModified types.String                `tfsdk:"time_modified"`
	Version      types.String                `tfsdk:"version"`
}

type imageDataSourceDigestModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

func (d *imageDataSource) Metadata(
	ctx context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = "oxide_image"
}

// Configure adds the provider configured client to the data source.
func (d *imageDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	_ *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *imageDataSource) Schema(
	ctx context.Context,
	req datasource.SchemaRequest,
	resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Retrieve information about a specified image.
`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the image.",
			},
			"project_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the project which contains the image.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"block_size": schema.Int64Attribute{
				Description: "Block size in bytes.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the image.",
			},
			"digest": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Hash of the image contents, if applicable.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Digest type.",
						Computed:    true,
					},
					"value": schema.StringAttribute{
						Description: "Digest type value.",
						Computed:    true,
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the image.",
			},
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the project which contains the image.",
			},
			"os": schema.StringAttribute{
				Computed:    true,
				Description: "OS image distribution.",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the image in bytes.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this image was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this image was last modified.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the OS.",
			},
		},
	}
}

func (d *imageDataSource) Read(
	ctx context.Context,
	req datasource.ReadRequest,
	resp *datasource.ReadResponse,
) {
	var state imageDataSourceModel

	// Read Terraform configuration data into the model
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

	params := oxide.ImageViewParams{
		Image:   oxide.NameOrId(state.Name.ValueString()),
		Project: oxide.NameOrId(state.ProjectName.ValueString()),
	}
	image, err := d.client.ImageView(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read image:",
			"API error: "+err.Error(),
		)
		return
	}
	tflog.Trace(
		ctx,
		fmt.Sprintf("read image with ID: %v", image.Id),
		map[string]any{"success": true},
	)

	// Map response body to model
	state.BlockSize = types.Int64Value(int64(image.BlockSize))
	state.Description = types.StringValue(image.Description)
	state.ID = types.StringValue(image.Id)
	state.ProjectID = types.StringValue(image.ProjectId)
	state.Name = types.StringValue(string(image.Name))
	state.OS = types.StringValue(image.Os)
	state.Size = types.Int64Value(int64(image.Size))
	state.TimeCreated = types.StringValue(image.TimeCreated.String())
	state.TimeModified = types.StringValue(image.TimeModified.String())
	state.Version = types.StringValue(image.Version)

	if image.Digest.Value != nil {
		sha256, ok := image.Digest.AsSha256()
		if !ok {
			resp.Diagnostics.AddError(
				"Unexpected digest type",
				fmt.Sprintf("Expected sha256 digest, got: %s", image.Digest.Type()),
			)
			return
		}
		digestState := imageDataSourceDigestModel{
			Type:  types.StringValue(string(image.Digest.Type())),
			Value: types.StringValue(sha256.Value),
		}
		state.Digest = &digestState
	}

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
