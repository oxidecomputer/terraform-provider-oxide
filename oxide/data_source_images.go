// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

func imagesDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: imagesDataSourceRead,
		Schema:      newImagesDataSourceSchema(),
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func newImagesDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:        schema.TypeString,
			Description: "ID of the project that contains the images.",
			Required:    true,
		},
		"images": {
			Computed:    true,
			Type:        schema.TypeList,
			Description: "A list of all images belonging to a project",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"block_size": {
						Type:        schema.TypeInt,
						Description: "Block size in bytes.",
						Computed:    true,
					},
					"description": {
						Type:        schema.TypeString,
						Description: "Description of the image.",
						Computed:    true,
					},
					"digest": {
						Type:        schema.TypeMap,
						Description: "Hash of the image contents, if applicable.",
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"os": {
						Type:        schema.TypeString,
						Description: "OS image distribution.",
						Computed:    true,
					},
					"id": {
						Type:        schema.TypeString,
						Description: "Unique, immutable, system-controlled identifier for the image.",
						Computed:    true,
					},
					"name": {
						Type:        schema.TypeString,
						Description: "Name of the image.",
						Computed:    true,
					},
					"size": {
						Type:        schema.TypeInt,
						Description: "Size of the image in bytes.",
						Computed:    true,
					},
					"time_created": {
						Type:        schema.TypeString,
						Description: "Timestamp of when this image was created.",
						Computed:    true,
					},
					"time_modified": {
						Type:        schema.TypeString,
						Description: "Timestamp of when this image was last modified.",
						Computed:    true,
					},
					"url": {
						Type:        schema.TypeString,
						Description: "URL source of this image, if any.",
						Computed:    true,
					},
					"version": {
						Type:        schema.TypeString,
						Description: "Version of the OS.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func imagesDataSourceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	projectId := d.Get("project_id").(string)

	// TODO: It would be preferable to us the client.Images.GlobalListAllPages method instead.
	// Unfortunately, currently that method has a bug where it returns twice as many results
	// as there are in reality. For now I'll use the List method with a limit of 1,000,000 results.
	// Seems unlikely anyone will have more than one billion Images.
	params := oxideSDK.ImageListParams{
		Project: oxideSDK.NameOrId(projectId),
		Limit:   1000000000,
		SortBy:  oxideSDK.NameOrIdSortModeIdAscending,
	}
	result, err := client.ImageList(params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(schema.HashString(time.Now().String())))

	if err := ImagesToState(d, result); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ImagesToState(d *schema.ResourceData, images *oxideSDK.ImageResultsPage) error {
	if images == nil {
		return nil
	}

	var result = make([]interface{}, 0, len(images.Items))
	for _, image := range images.Items {
		var m = make(map[string]interface{})

		m["block_size"] = image.BlockSize
		m["description"] = image.Description
		m["os"] = image.Os
		m["id"] = image.Id
		m["name"] = image.Name
		m["size"] = image.Size
		m["time_created"] = image.TimeCreated.String()
		m["time_modified"] = image.TimeModified.String()
		m["url"] = image.Url
		m["version"] = image.Version

		if digestFlattened := flattenDigest(image.Digest); digestFlattened != nil {
			m["digest"] = digestFlattened
		}

		result = append(result, m)

		if len(result) > 0 {
			if err := d.Set("images", result); err != nil {
				return err
			}
		}
	}

	return nil
}

func flattenDigest(digest oxideSDK.Digest) map[string]interface{} {
	var result = make(map[string]interface{})
	if digest.Type != "" {
		result[string(digest.Type)] = digest.Value
	}
	return result
}
