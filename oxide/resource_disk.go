// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = (*diskResource)(nil)
	_ resource.ResourceWithConfigure = (*diskResource)(nil)
)

// NewDiskResource is a helper function to simplify the provider implementation.
func NewDiskResource() resource.Resource {
	return &diskResource{}
}

// diskResource is the resource implementation.
type diskResource struct {
	client *oxideSDK.Client
}

type diskResourceModel struct {
	BlockSize    types.Int64  `tfsdk:"block_size"`
	Description  types.String `tfsdk:"description"`
	DevicePath   types.String `tfsdk:"device_path"`
	DiskSource   types.Map    `tfsdk:"disk_source"`
	ID           types.String `tfsdk:"id"`
	ImageID      types.String `tfsdk:"image_id"`
	Name         types.String `tfsdk:"name"`
	ProjectID    types.String `tfsdk:"project_id"`
	Size         types.Int64  `tfsdk:"size"`
	State        types.Object `tfsdk:"state"`
	SnapshotID   types.String `tfsdk:"snapshot_id"`
	TimeCreated  types.String `tfsdk:"time_created"`
	TimeModified types.String `tfsdk:"time_modified"`
}

type diskResourceStateModel struct {
	State    types.String `tfsdk:"state"`
	Instance types.String `tfsdk:"instance"`
}

// Metadata returns the resource type name.
func (r *diskResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_disk"
}

// Configure adds the provider configured client to the data source.
func (r *diskResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxideSDK.Client)
}

// Schema defines the schema for the resource.
func (r *diskResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the project that will contain the disk.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the disk.",
			},
			"disk_source": schema.MapAttribute{
				Required:    true,
				Description: "Source of a disk. Can be one of `blank = <block_size>`, `image = <image_id>`, `global_image = <image_id>`, or `snapshot = <snapshot_id>`.",
				ElementType: types.StringType,
			},
			"size": schema.Int64Attribute{
				Required:    true,
				Description: "Size of the disk in bytes.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description for the disk.",
			},
			"block_size": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of blocks in bytes.",
			},
			"device_path": schema.StringAttribute{
				Computed:    true,
				Description: "Path of the disk.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the image.",
			},
			"image_id": schema.StringAttribute{
				Computed:    true,
				Description: "Image ID of the disk source if applicable.",
			},
			"snapshot_id": schema.StringAttribute{
				Computed:    true,
				Description: "Snapshot ID of the disk source if applicable.",
			},

			"state": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "State of a Disk (primarily: attached or not).",
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						Description: "State of the disk.",
						Computed:    true,
					},
					"instance": schema.StringAttribute{
						Description: "Associated instance.",
						Computed:    true,
					},
				},
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this image was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this image was last modified.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *diskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan diskResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := oxideSDK.DiskCreateParams{
		Project: oxideSDK.NameOrId(plan.ProjectID.ValueString()),
		Body: &oxideSDK.DiskCreate{
			Description: plan.Description.ValueString(),
			Name:        oxideSDK.Name(plan.Name.ValueString()),
			Size:        oxideSDK.ByteCount(plan.Size.ValueInt64()),
		},
	}

	ds, err := newDiskSource(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse disk source:",
			err.Error(),
		)
		return
	}
	params.Body.DiskSource = ds

	disk, err := r.client.DiskCreate(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating disk",
			"API error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(disk.Id)
	plan.DevicePath = types.StringValue(disk.DevicePath)
	plan.BlockSize = types.Int64Value(int64(disk.BlockSize))
	plan.ImageID = types.StringValue(disk.ImageId)
	plan.SnapshotID = types.StringValue(disk.SnapshotId)
	plan.TimeCreated = types.StringValue(disk.TimeCreated.String())
	plan.TimeModified = types.StringValue(disk.TimeCreated.String())

	// Parse diskResourceStateModel into types.Object
	sm := diskResourceStateModel{
		State:    types.StringValue(string(disk.State.State)),
		Instance: types.StringValue(disk.State.Instance),
	}
	attributeTypes := map[string]attr.Type{
		"state":    types.StringType,
		"instance": types.StringType,
	}
	state, diags := types.ObjectValueFrom(ctx, attributeTypes, sm)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.State = state

	// Save plan into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *diskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state diskResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	disk, err := r.client.DiskView(oxideSDK.DiskViewParams{
		Disk: oxideSDK.NameOrId(state.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read disk:",
			"API error: "+err.Error(),
		)
		return
	}

	state.BlockSize = types.Int64Value(int64(disk.BlockSize))
	state.Description = types.StringValue(disk.Description)
	state.DevicePath = types.StringValue(disk.DevicePath)
	state.ID = types.StringValue(disk.Id)
	state.ImageID = types.StringValue(disk.ImageId)
	state.Name = types.StringValue(string(disk.Name))
	state.ProjectID = types.StringValue(disk.ProjectId)
	state.Size = types.Int64Value(int64(disk.Size))
	state.SnapshotID = types.StringValue(disk.SnapshotId)
	state.TimeCreated = types.StringValue(disk.TimeCreated.String())
	state.TimeModified = types.StringValue(disk.TimeCreated.String())

	// Parse diskResourceStateModel into types.Object
	sm := diskResourceStateModel{
		State:    types.StringValue(string(disk.State.State)),
		Instance: types.StringValue(disk.State.Instance),
	}
	attributeTypes := map[string]attr.Type{
		"state":    types.StringType,
		"instance": types.StringType,
	}
	diskState, diags := types.ObjectValueFrom(ctx, attributeTypes, sm)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.State = diskState

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *diskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating image",
		"the oxide API currently does not support updating images")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *diskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state diskResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Update this with detaching instance if state attached

	// Wait for disk to be detached before attempting to destroy.
	// TODO: For the time being there is no endpoint to detach disks without
	// knowing the Instance name first. The Disk get endpoint only retrieves
	// the attached instance ID, so we can't get the name from there.
	// This means that we cannot automatically detach disks here.
	// for a temporary workaround for the acceptance tests we will only check for a `detached`
	// status for 5 seconds and return an error otherwise.
	ch := make(chan error)
	go waitForDetachedDisk(r.client, oxideSDK.NameOrId(state.ID.ValueString()), ch)
	e := <-ch
	if e != nil {
		resp.Diagnostics.AddError(
			"Unable to delete disk:",
			"API error: "+e.Error(),
		)
		return
	}

	if err := r.client.DiskDelete(oxideSDK.DiskDeleteParams{
		Disk: oxideSDK.NameOrId(state.ID.ValueString()),
	}); err != nil {
		if !is404(err) {
			resp.Diagnostics.AddError(
				"Unable to delete disk:",
				"API error: "+err.Error(),
			)
			return
		}
	}
}

func newDiskSource(p diskResourceModel) (oxideSDK.DiskSource, error) {
	var ds = oxideSDK.DiskSource{}

	diskSource := p.DiskSource.Elements()
	if len(diskSource) > 1 {
		return ds, errors.New(
			"only one of blank = <block_size>, image = <image_id>, " +
				"global_image = <image_id>, or snapshot = <snapshot_id> can be set",
		)
	}

	if source, ok := diskSource["blank"]; ok {
		rawBs := source.String()
		blockSize, err := strconv.Unquote(rawBs)
		if err != nil {
			return ds, err
		}
		bs, err := strconv.Atoi(blockSize)
		if err != nil {
			return ds, err
		}
		ds = oxideSDK.DiskSource{
			BlockSize: oxideSDK.BlockSize(bs),
			Type:      oxideSDK.DiskSourceTypeBlank,
		}
	}

	if source, ok := diskSource["snapshot"]; ok {
		ds = oxideSDK.DiskSource{
			SnapshotId: source.String(),
			Type:       oxideSDK.DiskSourceTypeSnapshot,
		}
	}

	if source, ok := diskSource["image"]; ok {
		ds = oxideSDK.DiskSource{
			ImageId: source.String(),
			Type:    oxideSDK.DiskSourceTypeImage,
		}
	}

	if source, ok := diskSource["global_image"]; ok {
		ds = oxideSDK.DiskSource{
			ImageId: source.String(),
			Type:    oxideSDK.DiskSourceTypeGlobalImage,
		}
	}

	return ds, nil
}

func waitForDetachedDisk(client *oxideSDK.Client, diskID oxideSDK.NameOrId, ch chan error) {
	for start := time.Now(); time.Since(start) < (5 * time.Second); {
		resp, err := client.DiskView(oxideSDK.DiskViewParams{Disk: diskID})
		if err != nil {
			ch <- err
		}
		if resp.State.State == "detached" {
			break
		}
		time.Sleep(time.Second)
	}
	ch <- nil
}
