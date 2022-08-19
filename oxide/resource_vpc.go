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

func vpcResource() *schema.Resource {
	return &schema.Resource{
		Description:   "",
		Schema:        newVPCSchema(),
		CreateContext: createVPC,
		ReadContext:   readVPC,
		UpdateContext: updateVPC,
		DeleteContext: deleteVPC,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func newVPCSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"organization_name": {
			Type:        schema.TypeString,
			Description: "Name of the organization.",
			Required:    true,
		},
		"project_name": {
			Type:        schema.TypeString,
			Description: "Name of the project.",
			Required:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the VPC.",
			Required:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Description for the VPC.",
			Required:    true,
		},
		"dns_name": {
			Type:        schema.TypeString,
			Description: "DNS name of the VPC.",
			Required:    true,
		},
		"ipv6_prefix": {
			Type:        schema.TypeString,
			Description: "All IPv6 subnets created from this VPC must be taken from this range, which should be a unique local address in the range `fd00::/48`. The default VPC Subnet will have the first `/64` range from this prefix.",
			// TODO: For demo purposes this range will be generated only. When we move forward from demo stage
			// this value should be optional/computed
			Computed: true,
		},
		"id": {
			Type:        schema.TypeString,
			Description: "Unique, immutable, system-controlled identifier.",
			Computed:    true,
		},
		"project_id": {
			Type:        schema.TypeString,
			Description: "Unique, immutable, system-controlled identifier.",
			Computed:    true,
		},
		"system_router_id": {
			Type:        schema.TypeString,
			Description: "SystemRouterID is the ID for the system router where subnet default routes are registered.",
			Computed:    true,
		},
		"time_created": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this VPC was created.",
			Computed:    true,
		},
		"time_modified": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this VPC was last modified.",
			Computed:    true,
		},
	}
}

func createVPC(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	dnsName := d.Get("dns_name").(string)

	body := oxideSDK.VpcCreate{
		Description: description,
		Name:        oxideSDK.Name(name),
		DnsName:     oxideSDK.Name(dnsName),
	}

	resp, err := client.VpcCreate(oxideSDK.Name(orgName), oxideSDK.Name(projectName), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readVPC(ctx, d, meta)
}

func readVPC(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	vpcName := d.Get("name").(string)
	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)

	resp, err := client.VpcView(oxideSDK.Name(orgName), oxideSDK.Name(projectName), oxideSDK.Name(vpcName))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := vpcToState(d, resp); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateVPC(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	dnsName := d.Get("dns_name").(string)

	body := oxideSDK.VpcUpdate{
		Description: description,
		// We cannot change the name of the VPC as it is used as an identifier for
		// the update in the Put method. Changing it would make it impossible for
		// terraform to know which VPC to update.
		// Name:        name,
		DnsName: oxideSDK.Name(dnsName),
	}

	resp, err := client.VpcUpdate(oxideSDK.Name(orgName), oxideSDK.Name(projectName), oxideSDK.Name(name), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readVPC(ctx, d, meta)
}

func deleteVPC(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	vpcName := d.Get("name").(string)
	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)

	if err := client.VpcDelete(oxideSDK.Name(orgName), oxideSDK.Name(projectName), oxideSDK.Name(vpcName)); err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func vpcToState(d *schema.ResourceData, vpc *oxideSDK.Vpc) error {
	if err := d.Set("description", vpc.Description); err != nil {
		return err
	}
	if err := d.Set("dns_name", vpc.DnsName); err != nil {
		return err
	}
	if err := d.Set("id", vpc.Id); err != nil {
		return err
	}
	if err := d.Set("ipv6_prefix", vpc.Ipv6Prefix); err != nil {
		return err
	}
	if err := d.Set("name", vpc.Name); err != nil {
		return err
	}
	if err := d.Set("project_id", vpc.ProjectId); err != nil {
		return err
	}
	if err := d.Set("system_router_id", vpc.SystemRouterId); err != nil {
		return err
	}
	if err := d.Set("time_created", vpc.TimeCreated.String()); err != nil {
		return err
	}
	if err := d.Set("time_modified", vpc.TimeModified.String()); err != nil {
		return err
	}

	return nil
}
