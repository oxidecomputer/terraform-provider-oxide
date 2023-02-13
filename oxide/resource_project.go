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
			customdiff.ValidateChange("organization", func(ctx context.Context, old, new, meta any) error {
				if old != nil && new.(string) != old.(string) {
					return fmt.Errorf("organization of project cannot be updated; please revert to: \"%s\"", old.(string))
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
		// TODO: Remove when organization endpoints are gone
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

	resp, err := client.ProjectCreateV1(oxideSDK.NameOrId(orgName), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readProject(ctx, d, meta)
}

func readProject(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	projectId := d.Get("id").(string)

	resp, err := client.ProjectViewV1(oxideSDK.NameOrId(projectId), oxideSDK.NameOrId(""))
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

	return nil
}

func updateProject(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	projectId := d.Get("id").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)

	body := oxideSDK.ProjectUpdate{
		Description: description,
		Name:        oxideSDK.Name(name),
	}

	resp, err := client.ProjectUpdateV1(oxideSDK.NameOrId(projectId), oxideSDK.NameOrId(""), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readProject(ctx, d, meta)
}

func deleteProject(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	projectId := d.Get("id").(string)

	if err := client.ProjectDeleteV1(oxideSDK.NameOrId(projectId), oxideSDK.NameOrId("")); err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
