// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

var _ datasource.DataSource = (*imagesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*imagesDataSource)(nil)

type imagesDataSource struct {
	client *oxideSDK.Client
}

type imagesDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Images    []imageModel `tfsdk:"images"`
}

type imageModel struct {
	BlockSize    types.Int64   `tfsdk:"block_size"`
	Description  types.String  `tfsdk:"description"`
	Digest       []digestModel `tfsdk:"digest"`
	ID           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	OS           types.String  `tfsdk:"os"`
	Size         types.Int64   `tfsdk:"size"`
	TimeCreated  types.String  `tfsdk:"time_created"`
	TimeModified types.String  `tfsdk:"time_modified"`
	URL          types.String  `tfsdk:"url"`
	Version      types.String  `tfsdk:"version"`
}

type digestModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

// NewImagesDataSource initialises an images datasource
func NewImagesDataSource() datasource.DataSource {
	return &imagesDataSource{}
}

func (d *imagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "oxide_images"
}

// Configure adds the provider configured client to the data source.
func (d *imagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*oxideSDK.Client)
}

func (d *imagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the project which contains the images",
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
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
						"digest": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Hash of the image contents, if applicable.",
							NestedObject: schema.NestedAttributeObject{
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
						"url": schema.StringAttribute{
							Computed:    true,
							Description: "URL source of this image, if any.",
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

	// TODO: It would be preferable to us the client.Images.ListAllPages method instead.
	// Unfortunately, currently that method has a bug where it returns twice as many results
	// as there are in reality. For now I'll use the List method with a limit of 1,000,000,000 results.
	// Seems unlikely anyone will have more than one billion images.
	params := oxideSDK.ImageListParams{
		Project: oxideSDK.NameOrId(state.ProjectID.ValueString()),
		Limit:   1000000000,
		SortBy:  oxideSDK.NameOrIdSortModeIdAscending,
	}
	images, err := d.client.ImageList(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read images:",
			err.Error(),
		)
		return
	}

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
			URL:          types.StringValue(image.Url),
			Version:      types.StringValue(image.Version),
		}

		digestState := digestModel{
			Type:  types.StringValue(string(image.Digest.Type)),
			Value: types.StringValue(image.Digest.Value),
		}

		imageState.Digest = append(imageState.Digest, digestState)

		state.Images = append(state.Images, imageState)
	}

	// Save state into Terraform state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

//func imagesDataSource() *schema.Resource {
//	return &schema.Resource{
//		ReadContext: imagesDataSourceRead,
//		Schema:      newImagesDataSourceSchema(),
//		Timeouts: &schema.ResourceTimeout{
//			Default: schema.DefaultTimeout(5 * time.Minute),
//		},
//	}
//}

//func newImagesDataSourceSchema() map[string]*schema.Schema {
//	return map[string]*schema.Schema{
//		"image_id": {
//			Type:        schema.TypeString,
//			Description: "ID of the image that contains the images.",
//			Required:    true,
//		},
//		"images": {
//			Computed:    true,
//			Type:        schema.TypeList,
//			Description: "A list of all images belonging to a image",
//			Elem: &schema.Resource{
//				Schema: map[string]*schema.Schema{
//					"block_size": {
//						Type:        schema.TypeInt,
//						Description: "Block size in bytes.",
//						Computed:    true,
//					},
//					"description": {
//						Type:        schema.TypeString,
//						Description: "Description of the image.",
//						Computed:    true,
//					},
//					"digest": {
//						Type:        schema.TypeMap,
//						Description: "Hash of the image contents, if applicable.",
//						Computed:    true,
//						Elem: &schema.Schema{
//							Type: schema.TypeString,
//						},
//					},
//					"os": {
//						Type:        schema.TypeString,
//						Description: "OS image distribution.",
//						Computed:    true,
//					},
//					"id": {
//						Type:        schema.TypeString,
//						Description: "Unique, immutable, system-controlled identifier for the image.",
//						Computed:    true,
//					},
//					"name": {
//						Type:        schema.TypeString,
//						Description: "Name of the image.",
//						Computed:    true,
//					},
//					"size": {
//						Type:        schema.TypeInt,
//						Description: "Size of the image in bytes.",
//						Computed:    true,
//					},
//					"time_created": {
//						Type:        schema.TypeString,
//						Description: "Timestamp of when this image was created.",
//						Computed:    true,
//					},
//					"time_modified": {
//						Type:        schema.TypeString,
//						Description: "Timestamp of when this image was last modified.",
//						Computed:    true,
//					},
//					"url": {
//						Type:        schema.TypeString,
//						Description: "URL source of this image, if any.",
//						Computed:    true,
//					},
//					"version": {
//						Type:        schema.TypeString,
//						Description: "Version of the OS.",
//						Computed:    true,
//					},
//				},
//			},
//		},
//	}
//}
//
//func imagesDataSourceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
//	client := meta.(*oxideSDK.Client)
//	imageId := d.Get("image_id").(string)
//
// TODO: It would be preferable to us the client.Images.GlobalListAllPages method instead.
// Unfortunately, currently that method has a bug where it returns twice as many results
// as there are in reality. For now I'll use the List method with a limit of 1,000,000 results.
// Seems unlikely anyone will have more than one billion Images.
//	params := oxideSDK.ImageListParams{
//		Image: oxideSDK.NameOrId(imageId),
//		Limit:   1000000000,
//		SortBy:  oxideSDK.NameOrIdSortModeIdAscending,
//	}
//	result, err := client.ImageList(params)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	d.SetId(strconv.Itoa(schema.HashString(time.Now().String())))
//
//	if err := ImagesToState(d, result); err != nil {
//		return diag.FromErr(err)
//	}
//
//	return nil
//}

//func ImagesToState(d *schema.ResourceData, images *oxideSDK.ImageResultsPage) error {
//	if images == nil {
//		return nil
//	}
//
//	var result = make([]interface{}, 0, len(images.Items))
//	for _, image := range images.Items {
//		var m = make(map[string]interface{})
//
//		m["block_size"] = image.BlockSize
//		m["description"] = image.Description
//		m["os"] = image.Os
//		m["id"] = image.Id
//		m["name"] = image.Name
//		m["size"] = image.Size
//		m["time_created"] = image.TimeCreated.String()
//		m["time_modified"] = image.TimeModified.String()
//		m["url"] = image.Url
//		m["version"] = image.Version
//
//		if digestFlattened := flattenDigest(image.Digest); digestFlattened != nil {
//			m["digest"] = digestFlattened
//		}
//
//		result = append(result, m)
//
//		if len(result) > 0 {
//			if err := d.Set("images", result); err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}
//
//func flattenDigest(digest oxideSDK.Digest) map[string]interface{} {
//	var result = make(map[string]interface{})
//	if digest.Type != "" {
//		result[string(digest.Type)] = digest.Value
//	}
//	return result
//}
