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

func projectResource() *schema.Resource {
	return &schema.Resource{
		Description:   "",
		Schema:        newProjectSchema(),
		CreateContext: createProject,
		ReadContext:   readProject,
		UpdateContext: updateProject,
		DeleteContext: deleteProject,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			// TODO: When there is an API to update projects by ID remove this check to allow name changes
			customdiff.ValidateChange("name", func(ctx context.Context, old, new, meta any) error {
				if old.(string) != "" && new.(string) != old.(string) {
					return fmt.Errorf("name of project cannot be updated via Terraform; please revert to: \"%s\"", old.(string))
				}
				return nil
			}),
			customdiff.ValidateChange("organization", func(ctx context.Context, old, new, meta any) error {
				if old != nil && new.(string) != old.(string) {
					return fmt.Errorf("organization of IP pool cannot be updated; please revert to: \"%s\"", old.(string))
				}
				return nil
			}),
		),
	}
}

func newProjectSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the project.",
			Required:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Description for the project.",
			Required:    true,
		},
		"organization_name": {
			Type:        schema.TypeString,
			Description: "Name of the organization.",
			Required:    true,
		},
		"id": {
			Type:        schema.TypeString,
			Description: "Unique, immutable, system-controlled identifier of the project.",
			Computed:    true,
		},
		"organization_id": {
			Type:        schema.TypeString,
			Description: "ID of the organization.",
			Computed:    true,
		},
		"time_created": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this project was created.",
			Computed:    true,
		},
		"time_modified": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this project was last modified.",
			Computed:    true,
		},
	}
}

func createProject(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	orgName := d.Get("organization_name").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)

	body := oxideSDK.ProjectCreate{
		Description: description,
		Name:        oxideSDK.Name(name),
	}

	resp, err := client.ProjectCreate(oxideSDK.Name(orgName), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readProject(ctx, d, meta)
}

func readProject(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	projectId := d.Get("id").(string)

	resp, err := client.ProjectViewById(projectId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", resp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("organization_id", resp.OrganizationId); err != nil {
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

	return nil
}

func updateProject(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	orgName := d.Get("organization_name").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)

	body := oxideSDK.ProjectUpdate{
		Description: description,
		// We cannot change the name of the project as it is used as an identifier for
		// the update in the Put method. Changing it would make it impossible for
		// terraform to know which project to update.
		// Name:        name,
	}

	resp, err := client.ProjectUpdate(oxideSDK.Name(orgName), oxideSDK.Name(name), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readProject(ctx, d, meta)
}

func deleteProject(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	orgName := d.Get("organization_name").(string)
	name := d.Get("name").(string)

	if err := client.ProjectDelete(oxideSDK.Name(orgName), oxideSDK.Name(name)); err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
