// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
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
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	ranges := d.Get("ranges").([]interface{})

	body := oxideSDK.IpPoolCreate{
		Description: description,
		Name:        oxideSDK.Name(name),
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
			params := oxideSDK.IpPoolRangeAddParams{Pool: oxideSDK.NameOrId(resp.Id)}
			_, err := client.IpPoolRangeAdd(params, &r)
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
	ipPoolId := d.Get("id").(string)

	params := oxideSDK.IpPoolViewParams{Pool: oxideSDK.NameOrId(ipPoolId)}
	resp, err := client.IpPoolView(params)
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

	listParams := oxideSDK.IpPoolRangeListParams{
		Pool:  oxideSDK.NameOrId(ipPoolId),
		Limit: 1000000000,
	}
	resp2, err := client.IpPoolRangeList(listParams)
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
	ipPoolId := d.Get("id").(string)

	params := oxideSDK.IpPoolUpdateParams{
		Pool: oxideSDK.NameOrId(ipPoolId),
	}
	body := oxideSDK.IpPoolUpdate{
		Description: description,
		Name:        oxideSDK.Name(name),
	}
	resp, err := client.IpPoolUpdate(params, &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readIpPool(ctx, d, meta)
}

func deleteIpPool(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	ipPoolId := d.Get("id").(string)

	params := oxideSDK.IpPoolRangeListParams{
		Pool:  oxideSDK.NameOrId(ipPoolId),
		Limit: 1000000000,
	}
	// Remove all IP pool ranges first
	resp, err := client.IpPoolRangeList(params)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, item := range resp.Items {
		var ipRange oxideSDK.IpRange
		rs := item.Range.(map[string]interface{})
		if isIPv4(rs["first"].(string)) {
			ipRange = oxideSDK.Ipv4Range{
				First: rs["first"].(string),
				Last:  rs["last"].(string),
			}
		} else if isIPv6(rs["first"].(string)) {
			ipRange = oxideSDK.Ipv6Range{
				First: rs["first"].(string),
				Last:  rs["last"].(string),
			}
		} else {
			// This should never happen as we are retrieving information from Nexus. If we do encounter
			// this error we have a huge problem.
			return diag.FromErr(
				fmt.Errorf(
					"the value %s retrieved from Nexus is neither a valid IPv4 or IPv6. If you encounter this error please contact support",
					rs["first"].(string),
				),
			)
		}

		params := oxideSDK.IpPoolRangeRemoveParams{
			Pool: oxideSDK.NameOrId(ipPoolId),
		}
		if err := client.IpPoolRangeRemove(params, &ipRange); err != nil {
			return diag.FromErr(err)
		}
	}

	// Delete IP Pool once all ranges have been removed
	deleteParams := oxideSDK.IpPoolDeleteParams{Pool: oxideSDK.NameOrId(ipPoolId)}
	if err := client.IpPoolDelete(deleteParams); err != nil {
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

		var ipRange oxideSDK.IpRange

		if isIPv4(ipR["first_address"].(string)) {
			ipRange = oxideSDK.Ipv4Range{
				First: ipR["first_address"].(string),
				Last:  ipR["last_address"].(string),
			}
		} else if isIPv6(ipR["first_address"].(string)) {
			ipRange = oxideSDK.Ipv6Range{
				First: ipR["first_address"].(string),
				Last:  ipR["last_address"].(string),
			}
		} else {
			return nil, fmt.Errorf("%s is neither a valid IPv4 or IPv6", ipR["first_address"].(string))
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
			return nil, fmt.Errorf(
				"internal error: %v is not map[string]interface{}. Debugging content: %+v. If you hit this bug, please contact support",
				reflect.TypeOf(item.Range),
				item.Range,
			)
		}
		result = append(result, m)
	}

	return result, nil
}
