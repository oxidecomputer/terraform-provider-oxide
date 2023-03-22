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
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("ipv6_prefix", func(ctx context.Context, old, new, meta any) error {
				if old.(string) != "" && new.(string) != old.(string) {
					return fmt.Errorf("ipv6_prefix of VPC cannot be updated; please revert to: \"%s\"", old.(string))
				}
				return nil
			}),
		),
	}
}

func newVPCSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:        schema.TypeString,
			Description: "ID of the project that will contain the VPC.",
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
			Computed:    true,
			Optional:    true,
		},
		"id": {
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

	projectId := d.Get("project_id").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	dnsName := d.Get("dns_name").(string)
	ipv6Prefix := d.Get("ipv6_prefix").(string)

	body := oxideSDK.VpcCreate{
		Description: description,
		Name:        oxideSDK.Name(name),
		DnsName:     oxideSDK.Name(dnsName),
	}

	if ipv6Prefix != "" {
		body.Ipv6Prefix = oxideSDK.Ipv6Net(ipv6Prefix)
	}

	resp, err := client.VpcCreate(oxideSDK.NameOrId(projectId), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readVPC(ctx, d, meta)
}

func readVPC(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	vpcId := d.Get("id").(string)

	resp, err := client.VpcView(oxideSDK.NameOrId(vpcId), oxideSDK.NameOrId(""))
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

	description := d.Get("description").(string)
	name := d.Get("name").(string)
	dnsName := d.Get("dns_name").(string)
	vpcId := d.Get("id").(string)

	body := oxideSDK.VpcUpdate{
		Description: description,
		Name:        oxideSDK.Name(name),
		DnsName:     oxideSDK.Name(dnsName),
	}

	resp, err := client.VpcUpdate(oxideSDK.NameOrId(vpcId), oxideSDK.NameOrId(""), &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readVPC(ctx, d, meta)
}

func deleteVPC(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	vpcId := d.Get("id").(string)

	res, err := client.VpcSubnetList(
		1000000,
		"",
		"",
		oxideSDK.NameOrIdSortModeIdAscending,
		oxideSDK.NameOrId(vpcId),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if res != nil {
		for _, subnet := range res.Items {
			if err := client.VpcSubnetDelete(
				oxideSDK.NameOrId(subnet.Id),
				"",
				"",
			); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err := client.VpcDelete(oxideSDK.NameOrId(vpcId), ""); err != nil {
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
