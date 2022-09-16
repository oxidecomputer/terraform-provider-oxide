// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

func ipPoolResource() *schema.Resource {
	return &schema.Resource{
		Description:   "",
		Schema:        newIpPoolSchema(),
		CreateContext: createIpPool,
		ReadContext:   readIpPool,
		UpdateContext: updateIpPool,
		DeleteContext: deleteIpPool,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			// TODO: When there is an API to update IP pools by ID remove this check to allow name changes
			customdiff.ValidateChange("name", func(ctx context.Context, old, new, meta any) error {
				if old.(string) != "" && new.(string) != old.(string) {
					return fmt.Errorf("name of IP pool cannot be updated via Terraform; please revert to: \"%s\"", old.(string))
				}
				return nil
			}),
			customdiff.ValidateChange("organization", func(ctx context.Context, old, new, meta any) error {
				if old != nil && new.(string) != old.(string) {
					return fmt.Errorf("organization of IP pool cannot be updated; please revert to: \"%s\"", old.(string))
				}
				return nil
			}),
			customdiff.ValidateChange("project", func(ctx context.Context, old, new, meta any) error {
				if old != nil && new.(string) != old.(string) {
					return fmt.Errorf("project of IP pool cannot be updated; please revert to: \"%s\"", old.(string))
				}
				return nil
			}),
		),
	}
}

func newIpPoolSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the disk.",
			Required:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Description for the disk.",
			Required:    true,
		},
		"organization_name": {
			Type:        schema.TypeString,
			Description: "Name of the organization.",
			Optional:    true,
		},
		"project_name": {
			Type:        schema.TypeString,
			Description: "Name of the project.",
			Optional:    true,
		},
		"project_id": {
			Type:        schema.TypeString,
			Description: "Unique, immutable, system-controlled identifier of the project.",
			Computed:    true,
		},
		"time_created": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this disk was created.",
			Computed:    true,
		},
		"time_modified": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this disk was last modified.",
			Computed:    true,
		},
	}
}

func createIpPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)

	body := oxideSDK.IpPoolCreate{
		Description: description,
		Name:        oxideSDK.Name(name),
	}

	if orgName != "" {
		body.Organization = oxideSDK.Name(orgName)
	}

	if projectName != "" {
		body.Project = oxideSDK.Name(projectName)
	}

	resp, err := client.IpPoolCreate(&body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readIpPool(ctx, d, meta)
}

func readIpPool(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	ipPoolName := d.Get("name").(string)

	resp, err := client.IpPoolView(oxideSDK.Name(ipPoolName))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", resp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", resp.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("time_created", resp.TimeCreated.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("time_modified", resp.TimeModified.String()); err != nil {
		return diag.FromErr(err)
	}
	if resp.ProjectId != "" {
		if err := d.Set("project_id", resp.ProjectId); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func updateIpPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	description := d.Get("description").(string)
	name := d.Get("name").(string)

	body := oxideSDK.IpPoolUpdate{
		Description: description,
		// We cannot change the name of the IP pool as it is used as an identifier for
		// the update in the Put method. Changing it would make it impossible for
		// terraform to know which IP pool to update.
		// Name:        name,
	}

	resp, err := client.IpPoolUpdate(oxideSDK.Name(name), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readIpPool(ctx, d, meta)
}

func deleteIpPool(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	ipPoolName := d.Get("name").(string)

	if err := client.IpPoolDelete(oxideSDK.Name(ipPoolName)); err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
