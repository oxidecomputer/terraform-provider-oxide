// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

func imageResource() *schema.Resource {
	return &schema.Resource{
		Description:   "",
		Schema:        newImageSchema(),
		CreateContext: createImage,
		ReadContext:   readImage,
		UpdateContext: updateImage,
		DeleteContext: deleteImage,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("name", func(ctx context.Context, old, new, meta any) error {
				if old.(string) != "" && new.(string) != old.(string) {
					return errors.New("the oxide_image resource currently does not support updates")
				}
				return nil
			}),
			customdiff.ValidateChange("description", func(ctx context.Context, old, new, meta any) error {
				if old.(string) != "" && new.(string) != old.(string) {
					return errors.New("the oxide_image resource currently does not support updates")
				}
				return nil
			}),
			customdiff.ValidateChange("image_source", func(ctx context.Context, old, new, meta any) error {
				if old != nil && len(new.(map[string]interface{})) != len(old.(map[string]interface{})) && len(old.(map[string]interface{})) > 0 {
					return errors.New("the oxide_image resource currently does not support updates")
				}
				return nil
			}),
			customdiff.ValidateChange("os", func(ctx context.Context, old, new, meta any) error {
				if old.(string) != "" && new.(string) != old.(string) {
					return errors.New("the oxide_image resource currently does not support updates")
				}
				return nil
			}),
			customdiff.ValidateChange("version", func(ctx context.Context, old, new, meta any) error {
				if old.(string) != "" && new.(string) != old.(string) {
					return errors.New("the oxide_image resource currently does not support updates")
				}
				return nil
			}),
			customdiff.ValidateChange("block_size", func(ctx context.Context, old, new, meta any) error {
				if old.(int) != 0 && new.(int) != old.(int) {
					return errors.New("the oxide_image resource currently does not support updates")
				}
				return nil
			}),
		),
	}
}

func newImageSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:        schema.TypeString,
			Description: "ID of the project that will contain the image.",
			Required:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the image.",
			Required:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Description for the image.",
			Required:    true,
		},
		"image_source": {
			Type:        schema.TypeMap,
			Description: "Source of an image. Can be one of url=<URL> or snapshot=<snapshot_id>.",
			Required:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"os": {
			Type:        schema.TypeString,
			Description: "OS image distribution. Example: alpine",
			Required:    true,
		},
		"version": {
			Type:        schema.TypeString,
			Description: "OS image version. Example: 3.16.",
			Required:    true,
		},
		"block_size": {
			Type:        schema.TypeInt,
			Description: "Size of blocks in bytes.",
			Required:    true,
		},
		"digest": {
			Type:        schema.TypeString,
			Description: "Digest is hash of the image contents, if applicable.",
			Computed:    true,
		},
		"id": {
			Type:        schema.TypeString,
			Description: "Unique, immutable, system-controlled identifier of the image.",
			Computed:    true,
		},
		"size": {
			Type:        schema.TypeInt,
			Description: "Size is total size in bytes.",
			Computed:    true,
		},
		"url": {
			Type:        schema.TypeString,
			Description: "URL is URL source of this image, if any.",
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
	}
}

func createImage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	projectId := d.Get("project_id").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	bs := d.Get("block_size").(int)
	os := d.Get("os").(string)
	version := d.Get("version").(string)

	is, err := newImageSource(d)
	if err != nil {
		return diag.FromErr(err)
	}

	body := oxideSDK.ImageCreate{
		Description: description,
		Name:        oxideSDK.Name(name),
		BlockSize:   oxideSDK.BlockSize(bs),
		Os:          os,
		Version:     version,
		Source:      is,
	}

	resp, err := client.ImageCreateV1("", oxideSDK.NameOrId(projectId), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readImage(ctx, d, meta)
}

func readImage(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	imageId := d.Get("id").(string)

	resp, err := client.ImageViewV1(imageId, oxideSDK.NameOrId(""), oxideSDK.NameOrId(""))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := imageToState(d, resp); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateImage(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO: Currently there is no endpoint to update a image. Update this function when such endpoint exists
	return diag.FromErr(errors.New("the oxide_image resource currently does not support updates"))
}

func deleteImage(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	imageId := d.Get("id").(string)

	// NB: This endpoint is not implemented yet
	if err := client.ImageDeleteV1(oxideSDK.NameOrId(""), oxideSDK.NameOrId(""), oxideSDK.NameOrId(imageId)); err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func imageToState(d *schema.ResourceData, image *oxideSDK.Image) error {
	if err := d.Set("name", image.Name); err != nil {
		return err
	}
	if err := d.Set("description", image.Description); err != nil {
		return err
	}
	if err := d.Set("block_size", image.BlockSize); err != nil {
		return err
	}
	// TODO: Verify if it's necessary to show the type
	if err := d.Set("digest", image.Digest.Value); err != nil {
		return err
	}
	if err := d.Set("os", image.Os); err != nil {
		return err
	}
	if err := d.Set("version", image.Version); err != nil {
		return err
	}
	if err := d.Set("id", image.Id); err != nil {
		return err
	}
	if err := d.Set("size", image.Size); err != nil {
		return err
	}
	if err := d.Set("url", image.Url); err != nil {
		return err
	}
	if err := d.Set("time_created", image.TimeCreated.String()); err != nil {
		return err
	}
	if err := d.Set("time_modified", image.TimeModified.String()); err != nil {
		return err
	}

	return nil

}

func newImageSource(d *schema.ResourceData) (oxideSDK.ImageSource, error) {
	var is = oxideSDK.ImageSource{}

	imageSource := d.Get("image_source").(map[string]interface{})
	if len(imageSource) > 1 {
		return is, errors.New(
			"only one of url=<URL>, or snapshot=<snapshot_id> can be set",
		)
	}

	if source, ok := imageSource["url"]; ok {
		is = oxideSDK.ImageSource{
			Url:  source.(string),
			Type: oxideSDK.ImageSourceTypeUrl,
		}
	}

	if source, ok := imageSource["snapshot"]; ok {
		is = oxideSDK.ImageSource{
			Id:   source.(string),
			Type: oxideSDK.ImageSourceTypeSnapshot,
		}
	}

	if _, ok := imageSource["you_can_boot_anything_as_long_as_its_alpine"]; ok {
		is = oxideSDK.ImageSource{
			Type: oxideSDK.ImageSourceTypeYouCanBootAnythingAsLongAsItsAlpine,
		}
	}

	return is, nil
}
