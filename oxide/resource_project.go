// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	description := d.Get("description").(string)
	name := d.Get("name").(string)

	params := oxideSDK.ProjectCreateParams{
		Body: &oxideSDK.ProjectCreate{
			Description: description,
			Name:        oxideSDK.Name(name),
		},
	}

	resp, err := client.ProjectCreate(params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readProject(ctx, d, meta)
}

func readProject(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	projectId := d.Get("id").(string)

	params := oxideSDK.ProjectViewParams{Project: oxideSDK.NameOrId(projectId)}
	resp, err := client.ProjectView(params)
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

	params := oxideSDK.ProjectUpdateParams{
		Project: oxideSDK.NameOrId(projectId),
		Body: &oxideSDK.ProjectUpdate{
			Description: description,
			Name:        oxideSDK.Name(name),
		},
	}
	resp, err := client.ProjectUpdate(params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readProject(ctx, d, meta)
}

func deleteProject(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	projectId := d.Get("id").(string)

	params := oxideSDK.ProjectDeleteParams{Project: oxideSDK.NameOrId(projectId)}
	if err := client.ProjectDelete(params); err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
