// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"errors"

	//"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*imageResource)(nil)
	_ resource.ResourceWithConfigure = (*imageResource)(nil)
)

// NewImageResource is a helper function to simplify the provider implementation.
func NewImageResource() resource.Resource {
	return &imageResource{}
}

// imageResource is the resource implementation.
type imageResource struct {
	client *oxideSDK.Client
}

type imageResourceModel struct {
	BlockSize   types.Int64  `tfsdk:"block_size"`
	Description types.String `tfsdk:"description"`
	//	Digest       []imageResourceDigestModel `tfsdk:"digest"`
	ID           types.String `tfsdk:"id"`
	ImageSource  types.Map    `tfsdk:"image_source"`
	Name         types.String `tfsdk:"name"`
	OS           types.String `tfsdk:"os"`
	ProjectID    types.String `tfsdk:"project_id"`
	Size         types.Int64  `tfsdk:"size"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
	URL          types.String `tfsdk:"url"`
	Version      types.String `tfsdk:"version"`
}

//type imageResourceDigestModel struct {
//	Type  types.String `tfsdk:"type"`
//	Value types.String `tfsdk:"value"`
//}

// Metadata returns the resource type name.
func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_image"
}

// Configure adds the provider configured client to the data source.
func (r *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxideSDK.Client)
}

// Schema defines the schema for the resource.
func (r *imageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the project that will contain the image.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the image.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description for the image.",
			},
			"image_source": schema.MapAttribute{
				Required:    true,
				Description: "Description for the image.",
				ElementType: types.StringType,
			},
			"os": schema.StringAttribute{
				Required:    true,
				Description: "OS image distribution. Example: alpine",
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "OS image version. Example: 3.16.",
			},
			"block_size": schema.Int64Attribute{
				Required:    true,
				Description: "Size of blocks in bytes.",
			},
			//	"digest": schema.ListNestedAttribute{
			//		Computed:    true,
			//		Description: "Hash of the image contents, if applicable.",
			//		NestedObject: schema.NestedAttributeObject{
			//			Attributes: map[string]schema.Attribute{
			//				"type": schema.StringAttribute{
			//					Description: "Digest type.",
			//					Computed:    true,
			//				},
			//				"value": schema.StringAttribute{
			//					Description: "Digest type value.",
			//					Computed:    true,
			//				},
			//			},
			//		},
			//	},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the image.",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "Total size in bytes.",
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
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := oxideSDK.ImageCreateParams{
		Project: oxideSDK.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxideSDK.ImageCreate{
			Description: plan.Description.ValueString(),
			Name:        oxideSDK.Name(plan.Name.ValueString()),
			BlockSize:   oxideSDK.BlockSize(plan.BlockSize.ValueInt64()),
			Os:          plan.OS.ValueString(),
			Version:     plan.Version.ValueString(),
		},
	}

	is, err := newImageSource(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse image source:",
			err.Error(),
		)
		return
	}
	params.Body.Source = is

	image, err := r.client.ImageCreate(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating image",
			"API error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(image.Id)
	plan.Size = types.Int64Value(int64(image.Size))
	plan.TimeCreated = types.StringValue(image.TimeCreated.String())
	plan.TimeModified = types.StringValue(image.TimeCreated.String())
	plan.URL = types.StringValue(image.Url)
	plan.Version = types.StringValue(image.Version)
	//	digestState := imageResourceDigestModel{
	//		Type:  types.StringValue(string(image.Digest.Type)),
	//		Value: types.StringValue(image.Digest.Value),
	//	}
	//	plan.Digest = append(plan.Digest, digestState)

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	image, err := r.client.ImageView(oxideSDK.ImageViewParams{
		Image: oxideSDK.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read image:",
			"API error: "+err.Error(),
		)
		return
	}

	state.BlockSize = types.Int64Value(int64(image.BlockSize))
	state.Description = types.StringValue(image.Description)
	state.ID = types.StringValue(image.Id)
	state.Name = types.StringValue(string(image.Name))
	state.OS = types.StringValue(image.Os)
	state.Size = types.Int64Value(int64(image.Size))
	state.TimeCreated = types.StringValue(image.TimeCreated.String())
	state.TimeModified = types.StringValue(image.TimeCreated.String())
	state.URL = types.StringValue(image.Url)
	state.Version = types.StringValue(image.Version)

	//	digestState := imageResourceDigestModel{
	//		Type:  types.StringValue(string(image.Digest.Type)),
	//		Value: types.StringValue(image.Digest.Value),
	//	}
	//	state.Digest = append(state.Digest, digestState)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating image",
		"the oxide API currently does not support updating images")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Error delete image",
		"the oxide API currently does not support deleting images")

	// TODO: Uncomment once image delete is enabled in the API
	//
	//	var state imageResourceModel
	//
	//	// Read Terraform prior state data into the model
	//	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	//	if resp.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	if err := r.client.ImageDelete(oxideSDK.ImageDeleteParams{
	//		Image: oxideSDK.NameOrId(state.ID.ValueString()),
	//	}); err != nil {
	//
	//		resp.Diagnostics.AddError(
	//			"Unable to read image:",
	//			"API error: "+err.Error(),
	//		)
	//		return
	//	}
}

func newImageSource(p imageResourceModel) (oxideSDK.ImageSource, error) {
	var is = oxideSDK.ImageSource{}

	imageSource := p.ImageSource.Elements()
	if len(imageSource) > 1 {
		return is, errors.New(
			"only one of url=<URL>, or snapshot=<snapshot_id> can be set",
		)
	}

	if source, ok := imageSource["url"]; ok {
		is = oxideSDK.ImageSource{
			Url:  source.String(),
			Type: oxideSDK.ImageSourceTypeUrl,
		}
	}

	if source, ok := imageSource["snapshot"]; ok {
		is = oxideSDK.ImageSource{
			Id:   source.String(),
			Type: oxideSDK.ImageSourceTypeSnapshot,
		}
	}

	// TODO: For testing only, remove before releasing
	if _, ok := imageSource["you_can_boot_anything_as_long_as_its_alpine"]; ok {
		is = oxideSDK.ImageSource{
			Type: oxideSDK.ImageSourceTypeYouCanBootAnythingAsLongAsItsAlpine,
		}
	}

	return is, nil
}
