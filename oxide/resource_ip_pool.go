// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
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
			// TODO: Enable adding and removing ranges. Figuring out best way forward for this
			customdiff.ValidateChange("ranges", func(ctx context.Context, old, new, meta any) error {
				if old != nil && len(new.([]interface{})) != len(old.([]interface{})) && len(old.([]interface{})) > 0 {
					return fmt.Errorf("IP pool ranges cannot be updated; please revert to previous configuration")
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
			Description: "Name of the IP pool.",
			Required:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Description for the IP pool.",
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
		"ranges": {
			Type:        schema.TypeList,
			Description: "A non-decreasing IPv4 or IPv6 address range, inclusive of both ends. The first address must be less than or equal to the last address.",
			Optional:    true,
			Elem:        newRangeResource(),
		},
		"id": {
			Type:        schema.TypeString,
			Description: "Unique, immutable, system-controlled identifier.",
			Computed:    true,
		},
		"project_id": {
			Type:        schema.TypeString,
			Description: "Unique, immutable, system-controlled identifier of the project.",
			Computed:    true,
		},
		"time_created": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this IP pool was created.",
			Computed:    true,
		},
		"time_modified": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this IP pool was last modified.",
			Computed:    true,
		},
	}
}

func newRangeResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			// TODO: Will likely have to remove this field as it complicates things for updates
			"ip_version": {
				Type:        schema.TypeString,
				Description: "IP version of the range. Accepted values are ipv4 and ipv6",
				Required:    true,
			},
			"first_address": {
				Type:        schema.TypeString,
				Description: "First address in the range.",
				Required:    true,
			},
			"last_address": {
				Type:        schema.TypeString,
				Description: "Last address in the range.",
				Required:    true,
			},
			"id": {
				Type:        schema.TypeString,
				Description: "Unique, immutable, system-controlled identifier.",
				Computed:    true,
			},
			"time_created": {
				Type:        schema.TypeString,
				Description: "Timestamp of when this range was created.",
				Computed:    true,
			},
		},
	}
}

func createIpPool(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	ranges := d.Get("ranges").([]interface{})

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

	if len(ranges) > 0 {
		ipRanges, err := newIpPoolRange(d)
		if err != nil {
			return diag.FromErr(err)
		}
		for _, r := range ipRanges {
			_, err := client.IpPoolRangeAdd(resp.Name, &r)
			// TODO: Remove when error from the API is more end user friendly
			if err != nil && strings.Contains(err.Error(), "data did not match any variant of untagged enum IpRange") {
				return diag.FromErr(fmt.Errorf("%+v is not an accepted IP range", r))
			}
			if err != nil {
				return diag.FromErr(fmt.Errorf("%v: %v", len(ipRanges), err))
			}
		}
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

	resp2, err := client.IpPoolRangeList(oxideSDK.Name(ipPoolName), 1000000000, "")
	if err != nil {
		return diag.FromErr(err)
	}

	ranges, err := ipPoolRangesToState(client, *resp2)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ranges", ranges); err != nil {
		return diag.FromErr(err)
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

	// TODO: Remove ranges first? Will find out if this is necessary when this endpoint is enabled

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

func newIpPoolRange(d *schema.ResourceData) ([]oxideSDK.IpRange, error) {
	rs := d.Get("ranges").([]interface{})

	var ipRanges []oxideSDK.IpRange

	for _, r := range rs {
		ipR := r.(map[string]interface{})

		if ipR["ip_version"].(string) != "ipv4" && ipR["ip_version"].(string) != "ipv6" {
			return nil, errors.New("ip_version must be one of \"ipv4\" or \"ipv6\"")
		}

		var ipRange oxideSDK.IpRange

		if ipR["ip_version"].(string) == "ipv4" {
			ipRange = oxideSDK.Ipv4Range{
				First: ipR["first_address"].(string),
				Last:  ipR["last_address"].(string),
			}
		}

		if ipR["ip_version"].(string) == "ipv6" {
			ipRange = oxideSDK.Ipv6Range{
				First: ipR["first_address"].(string),
				Last:  ipR["last_address"].(string),
			}
		}

		ipRanges = append(ipRanges, ipRange)
	}

	return ipRanges, nil
}

func ipPoolRangesToState(client *oxideSDK.Client, ipPoolRange oxideSDK.IpPoolRangeResultsPage) ([]interface{}, error) {
	items := ipPoolRange.Items
	var result = make([]interface{}, 0, len(items))
	for _, item := range items {
		var m = make(map[string]interface{})

		m["id"] = item.Id
		m["time_created"] = item.TimeCreated.String()

		// TODO: For the time being we are using interfaces for nested allOf within oneOf objects in
		// the OpenAPI spec. When we come up with a better approach this should be edited to reflect that.
		switch item.Range.(type) {
		case map[string]interface{}:
			rs := item.Range.(map[string]interface{})
			m["first_address"] = rs["first"]
			m["last_address"] = rs["last"]
		default:
			// Theoretically this should never happen. Just in case though!
			return nil, fmt.Errorf("internal error: %v is not map[string]interface{}. Debugging content: %+v. If you hit this bug, please contact support", reflect.TypeOf(item.Range), item.Range)
		}
		result = append(result, m)
	}

	return result, nil
}
