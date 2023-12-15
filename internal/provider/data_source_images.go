// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var (
	_ datasource.DataSource              = (*imagesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*imagesDataSource)(nil)
)

// NewImagesDataSource initialises an images datasource
func NewImagesDataSource() datasource.DataSource {
	return &imagesDataSource{}
}

type imagesDataSource struct {
	client *oxide.Client
}

type imagesDataSourceModel struct {
	ID        types.String   `tfsdk:"id"`
	ProjectID types.String   `tfsdk:"project_id"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
	Images    []imageModel   `tfsdk:"images"`
}

type imageModel struct {
	BlockSize    types.Int64      `tfsdk:"block_size"`
	Description  types.String     `tfsdk:"description"`
	Digest       imageDigestModel `tfsdk:"digest"`
	ID           types.String     `tfsdk:"id"`
	Name         types.String     `tfsdk:"name"`
	OS           types.String     `tfsdk:"os"`
	Size         types.Int64      `tfsdk:"size"`
	TimeCreated  types.String     `tfsdk:"time_created"`
	TimeModified types.String     `tfsdk:"time_modified"`
	Version      types.String     `tfsdk:"version"`
}

type imageDigestModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

func (d *imagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_images"
}

// Configure adds the provider configured client to the data source.
func (d *imagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxide.Client)
}

func (d *imagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the project which contains the images",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"images": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
						"os": schema.StringAttribute{
							Computed:    true,
							Description: "OS image distribution.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the image.",
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
				},
			},
		},
	}
}

func (d *imagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state imagesDataSourceModel

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

	// TODO: It would be preferable to us the client.Images.ListAllPages method instead.
	// Unfortunately, currently that method has a bug where it returns twice as many results
	// as there are in reality. For now I'll use the List method with a limit of 1,000,000,000 results.
	// Seems unlikely anyone will have more than one billion images.
	params := oxide.ImageListParams{
		Project: oxide.NameOrId(state.ProjectID.ValueString()),
		Limit:   1000000000,
		SortBy:  oxide.NameOrIdSortModeIdAscending,
	}
	images, err := d.client.ImageList(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read images:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read all images from project: %v", state.ProjectID.ValueString()), map[string]any{"success": true})

	// Set a unique ID for the datasource payload
	state.ID = types.StringValue(uuid.New().String())

	// Map response body to model
	for _, image := range images.Items {
		imageState := imageModel{
			BlockSize:    types.Int64Value(int64(image.BlockSize)),
			Description:  types.StringValue(image.Description),
			ID:           types.StringValue(image.Id),
			Name:         types.StringValue(string(image.Name)),
			OS:           types.StringValue(image.Os),
			Size:         types.Int64Value(int64(image.Size)),
			TimeCreated:  types.StringValue(image.TimeCreated.String()),
			TimeModified: types.StringValue(image.TimeCreated.String()),
			Version:      types.StringValue(image.Version),
		}

		digestState := imageDigestModel{
			Type:  types.StringValue(string(image.Digest.Type)),
			Value: types.StringValue(image.Digest.Value),
		}

		imageState.Digest = digestState

		state.Images = append(state.Images, imageState)
	}

	// Save state into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
