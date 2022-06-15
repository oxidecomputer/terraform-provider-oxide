// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go"
)

func diskResource() *schema.Resource {
	return &schema.Resource{
		Description:   "",
		Schema:        newDiskSchema(),
		CreateContext: createDisk,
		ReadContext:   readDisk,
		UpdateContext: updateDisk,
		DeleteContext: deleteDisk,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func newDiskSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"organization_name": {
			Type:        schema.TypeString,
			Description: "Name of the organization",
			Required:    true,
		},
		"project_name": {
			Type:        schema.TypeString,
			Description: "Name of the project",
			Required:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the disk",
			Required:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Description for the disk",
			Required:    true,
		},
		"disk_source": {
			Type:        schema.TypeMap,
			Description: "Source of a disk. Can be one of blank=block_size, image=image_id, global_image=image_id, or snapshot=snapshot_id",
			Required:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"size": {
			Type:        schema.TypeInt,
			Description: "Size of the disk in bytes",
			Required:    true,
		},
		"block_size": {
			Type:        schema.TypeInt,
			Description: "Size of blocks in bytes",
			Computed:    true,
		},
		"image_id": {
			Type:        schema.TypeString,
			Description: "Image ID of the disk source",
			Computed:    true,
		},
		"snapshot_id": {
			Type:        schema.TypeString,
			Description: "Snapshot ID of the disk source",
			Computed:    true,
		},
		"device_path": {
			Type:        schema.TypeString,
			Description: "Path of the disk",
			Computed:    true,
		},
		"id": {
			Type:        schema.TypeString,
			Description: "Immutable disk ID",
			Computed:    true,
		},
		"project_id": {
			Type:        schema.TypeString,
			Description: "Immutable project ID",
			Computed:    true,
		},
		"state": {
			Type:        schema.TypeList,
			Description: "State of the disk",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"state": {
						Type:     schema.TypeString,
						Computed: true,
					},

					"instance": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"time_created": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"time_modified": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func createDisk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	size := d.Get("size").(int)

	ds, err := newDiskSource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	body := oxideSDK.DiskCreate{
		Description: description,
		Name:        name,
		DiskSource:  ds,
		Size:        oxideSDK.ByteCount(size),
	}

	resp, err := client.Disks.Create(orgName, projectName, &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return readDisk(ctx, d, meta)
}

func readDisk(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	diskName := d.Get("name").(string)
	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)

	resp, err := client.Disks.Get(diskName, orgName, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := diskToState(d, resp); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateDisk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO: Currently there is no endpoint to update a disk. This function will remain
	// as readonly until such endpoint exists.
	return readDisk(ctx, d, meta)
}

func deleteDisk(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	diskName := d.Get("name").(string)
	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)

	if err := client.Disks.Delete(diskName, orgName, projectName); err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func diskToState(d *schema.ResourceData, disk *oxideSDK.Disk) error {
	if err := d.Set("block_size", disk.BlockSize); err != nil {
		return err
	}
	if err := d.Set("image_id", disk.ImageID); err != nil {
		return err
	}
	if err := d.Set("snapshot_id", disk.SnapshotID); err != nil {
		return err
	}
	if err := d.Set("description", disk.Description); err != nil {
		return err
	}
	if err := d.Set("device_path", disk.DevicePath); err != nil {
		return err
	}
	if err := d.Set("id", disk.ID); err != nil {
		return err
	}
	if err := d.Set("name", disk.Name); err != nil {
		return err
	}
	if err := d.Set("project_id", disk.ProjectID); err != nil {
		return err
	}
	if err := d.Set("size", disk.Size); err != nil {
		return err
	}
	if err := d.Set("time_created", disk.TimeCreated.String()); err != nil {
		return err
	}
	if err := d.Set("time_modified", disk.TimeModified.String()); err != nil {
		return err
	}

	var result = make([]interface{}, 0, 2)
	var m = make(map[string]interface{})
	m["state"] = disk.State.State
	m["instance"] = disk.State.Instance
	result = append(result, m)
	if err := d.Set("state", result); err != nil {
		return err

	}

	return nil

}

func newDiskSource(d *schema.ResourceData) (oxideSDK.DiskSource, error) {
	var ds = oxideSDK.DiskSource{}

	diskSource := d.Get("disk_source").(map[string]interface{})
	if len(diskSource) > 1 {
		return ds, errors.New(
			"only one of blank=block_size, image=image_id, global_image=image_id, or snapshot=snapshot_id can be set",
		)
	}

	if source, ok := diskSource["blank"]; ok {
		rawBs := source.(string)
		bs, err := strconv.Atoi(rawBs)
		if err != nil {
			return ds, err
		}
		ds = oxideSDK.DiskSource{
			BlockSize: oxideSDK.BlockSize(bs),
			Type:      "blank",
		}
	}

	if source, ok := diskSource["image"]; ok {
		ds = oxideSDK.DiskSource{
			ImageID: source.(string),
			Type:    "image",
		}
	}

	if source, ok := diskSource["global_image"]; ok {
		ds = oxideSDK.DiskSource{
			ImageID: source.(string),
			Type:    "global_image",
		}
	}

	if source, ok := diskSource["snapshot"]; ok {
		ds = oxideSDK.DiskSource{
			SnapshotID: source.(string),
			Type:       "snapshot",
		}
	}

	return ds, nil
}
