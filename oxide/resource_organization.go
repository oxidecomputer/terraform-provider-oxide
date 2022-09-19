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

func organizationResource() *schema.Resource {
	return &schema.Resource{
		Description:   "",
		Schema:        newOrganizationSchema(),
		CreateContext: createOrganization,
		ReadContext:   readOrganization,
		UpdateContext: updateOrganization,
		DeleteContext: deleteOrganization,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			// TODO: When there is an API to update organizations by ID remove this check to allow name changes
			customdiff.ValidateChange("name", func(ctx context.Context, old, new, meta any) error {
				if old.(string) != "" && new.(string) != old.(string) {
					return fmt.Errorf("name of organization cannot be updated via Terraform; please revert to: \"%s\"", old.(string))
				}
				return nil
			}),
		),
	}
}

func newOrganizationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the organization.",
			Required:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Description for the organization.",
			Required:    true,
		},
		"id": {
			Type:        schema.TypeString,
			Description: "Unique, immutable, system-controlled identifier of the organization.",
			Computed:    true,
		},
		"time_created": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this organization was created.",
			Computed:    true,
		},
		"time_modified": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this organization was last modified.",
			Computed:    true,
		},
	}
}

func createOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	description := d.Get("description").(string)
	name := d.Get("name").(string)

	body := oxideSDK.OrganizationCreate{
		Description: description,
		Name:        oxideSDK.Name(name),
	}

	resp, err := client.OrganizationCreate(&body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readOrganization(ctx, d, meta)
}

func readOrganization(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	orgID := d.Get("id").(string)

	resp, err := client.OrganizationViewById(orgID)
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

func updateOrganization(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	description := d.Get("description").(string)
	name := d.Get("name").(string)

	body := oxideSDK.OrganizationUpdate{
		Description: description,
		// We cannot change the name of the organization as it is used as an identifier for
		// the update in the Put method. Changing it would make it impossible for
		// terraform to know which organization to update.
		// Name:        name,
	}

	resp, err := client.OrganizationUpdate(oxideSDK.Name(name), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readOrganization(ctx, d, meta)
}

func deleteOrganization(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	orgName := d.Get("name").(string)

	if err := client.OrganizationDelete(oxideSDK.Name(orgName)); err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
