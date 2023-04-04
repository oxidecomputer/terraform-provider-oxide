// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

func instanceResource() *schema.Resource {
	return &schema.Resource{
		Description:   "",
		Schema:        newInstanceSchema(),
		CreateContext: createInstance,
		ReadContext:   readInstance,
		UpdateContext: updateInstance,
		DeleteContext: deleteInstance,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func newInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:        schema.TypeString,
			Description: "ID of the project that will contain the instance.",
			Required:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Name of the instance.",
			Required:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Description for the instance.",
			Required:    true,
		},
		"host_name": {
			Type:        schema.TypeString,
			Description: "Host name of the instance.",
			Required:    true,
		},
		"memory": {
			Type:        schema.TypeInt,
			Description: "Instance memory in bytes.",
			Required:    true,
		},
		"ncpus": {
			Type:        schema.TypeInt,
			Description: "Number of CPUs allocated for this instance.",
			Required:    true,
		},
		"attach_to_disks": {
			Type:        schema.TypeList,
			Description: "Disks to be attached to this instance.",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"external_ips": {
			Type:        schema.TypeList,
			Description: "External IP addresses provided to this instance. List of IP pools from which to draw addresses.",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"network_interface": {
			Type:        schema.TypeList,
			Description: "Attaches network interfaces to an instance at the time the instance is created.",
			Optional:    true,
			Elem:        newNetworkInterfaceResource(),
		},
		"id": {
			Type:        schema.TypeString,
			Description: "Unique, immutable, system-controlled identifier.",
			Computed:    true,
		},
		"run_state": {
			Type:        schema.TypeString,
			Description: "Running state of an Instance (primarily: booted or stopped). This typically reflects whether it's starting, running, stopping, or stopped, but also includes states related to the instance's lifecycle.",
			Computed:    true,
		},
		"time_created": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this instance was created.",
			Computed:    true,
		},
		"time_modified": {
			Type:        schema.TypeString,
			Description: "Timestamp of when this instance was last modified.",
			Computed:    true,
		},
		"time_run_state_updated": {
			Type:        schema.TypeString,
			Description: "Timestamp of when the run state of this instance was last modified.",
			Computed:    true,
		},
	}
}

func newNetworkInterfaceResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Description: "Description for the network interface.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name for the network interface.",
				Required:    true,
			},
			"subnet_name": {
				Type:        schema.TypeString,
				Description: "Name of the VPC Subnet in which to create the network interface.",
				Required:    true,
			},
			"vpc_name": {
				Type:        schema.TypeString,
				Description: "Name of the VPC in which to create the network interface.",
				Required:    true,
			},
			"ip": {
				Type:        schema.TypeString,
				Description: "IP address for the network interface.",
				// TODO: For the purposes of this demo we will stick to
				// auto-assigned IP addresses. In the future we will want
				// this value to be computed/optional or required.
				Computed: true,
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Description: "ID of the VPC Subnet to which the interface belongs.",
				Computed:    true,
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Description: "ID of the VPC in which to which the interface belongs.",
				Computed:    true,
			},
		},
	}
}

func createInstance(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	projectId := d.Get("project_id").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	hostName := d.Get("host_name").(string)
	memory := d.Get("memory").(int)
	ncpus := d.Get("ncpus").(int)

	params := oxideSDK.InstanceCreateParams{
		Project: oxideSDK.NameOrId(projectId),
		Body: &oxideSDK.InstanceCreate{
			Description:       description,
			Name:              oxideSDK.Name(name),
			Hostname:          hostName,
			Memory:            oxideSDK.ByteCount(memory),
			Ncpus:             oxideSDK.InstanceCpuCount(ncpus),
			Disks:             newInstanceDiskAttach(d),
			ExternalIps:       newInstanceExternalIps(d),
			NetworkInterfaces: newNetworkInterface(d),
		},
	}

	resp, err := client.InstanceCreate(params)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	return readInstance(ctx, d, meta)
}

func readInstance(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	instanceId := d.Get("id").(string)

	params := oxideSDK.InstanceViewParams{Instance: oxideSDK.NameOrId(instanceId)}
	resp, err := client.InstanceView(params)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := instanceToState(d, resp); err != nil {
		return diag.FromErr(err)
	}

	nis := d.Get("network_interface").([]interface{})
	if len(nis) > 0 {
		nicParams := oxideSDK.InstanceNetworkInterfaceListParams{
			Instance: oxideSDK.NameOrId(instanceId),
			Limit:    1000000000,
			SortBy:   oxideSDK.NameOrIdSortModeNameAscending,
		}
		resp2, err := client.InstanceNetworkInterfaceList(nicParams)
		if err != nil {
			return diag.FromErr(err)
		}

		networkInterfaces, err := networkInterfaceToState(client, *resp2)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("network_interface", networkInterfaces); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func updateInstance(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO: Currently there is no endpoint to update an instance. Update this function when such endpoint exists
	return diag.FromErr(errors.New("the oxide_instance resource currently does not support updates"))
}

func deleteInstance(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	instanceId := d.Get("id").(string)

	params := oxideSDK.InstanceStopParams{Instance: oxideSDK.NameOrId(instanceId)}
	_, err := client.InstanceStop(params)
	if err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Wait for instance to be stopped before attempting to destroy
	ch := make(chan error)
	go waitForStoppedInstance(client, oxideSDK.NameOrId(instanceId), ch)
	e := <-ch
	if e != nil {
		return diag.FromErr(e)
	}

	delParams := oxideSDK.InstanceDeleteParams{Instance: oxideSDK.NameOrId(instanceId)}
	if err := client.InstanceDelete(delParams); err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func instanceToState(d *schema.ResourceData, instance *oxideSDK.Instance) error {
	if err := d.Set("name", instance.Name); err != nil {
		return err
	}
	if err := d.Set("description", instance.Description); err != nil {
		return err
	}
	if err := d.Set("host_name", instance.Hostname); err != nil {
		return err
	}
	if err := d.Set("memory", instance.Memory); err != nil {
		return err
	}
	if err := d.Set("ncpus", instance.Ncpus); err != nil {
		return err
	}
	if err := d.Set("id", instance.Id); err != nil {
		return err
	}
	if err := d.Set("project_id", instance.ProjectId); err != nil {
		return err
	}
	if err := d.Set("run_state", instance.RunState); err != nil {
		return err
	}
	if err := d.Set("time_created", instance.TimeCreated.String()); err != nil {
		return err
	}
	if err := d.Set("time_modified", instance.TimeModified.String()); err != nil {
		return err
	}
	if err := d.Set("time_run_state_updated", instance.TimeRunStateUpdated.String()); err != nil {
		return err
	}

	return nil
}

func waitForStoppedInstance(client *oxideSDK.Client, instanceId oxideSDK.NameOrId, ch chan error) {
	for {
		params := oxideSDK.InstanceViewParams{Instance: instanceId}
		resp, err := client.InstanceView(params)
		if err != nil {
			ch <- err
		}
		if resp.RunState == oxideSDK.InstanceStateStopped {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	ch <- nil
}

func newInstanceExternalIps(d *schema.ResourceData) []oxideSDK.ExternalIpCreate {
	var externalIps = []oxideSDK.ExternalIpCreate{}
	ips := d.Get("external_ips").([]interface{})

	if len(ips) < 1 {
		return externalIps
	}
	for _, ip := range ips {
		ds := oxideSDK.ExternalIpCreate{
			PoolName: oxideSDK.Name(ip.(string)),
			// TODO: Implement other types when these are supported.
			Type: oxideSDK.ExternalIpCreateTypeEphemeral,
		}

		externalIps = append(externalIps, ds)
	}

	return externalIps
}

func newInstanceDiskAttach(d *schema.ResourceData) []oxideSDK.InstanceDiskAttachment {
	var diskAttachement = []oxideSDK.InstanceDiskAttachment{}
	disks := d.Get("attach_to_disks").([]interface{})

	if len(disks) < 1 {
		return diskAttachement
	}
	for _, disk := range disks {
		ds := oxideSDK.InstanceDiskAttachment{
			Name: oxideSDK.Name(disk.(string)),
			Type: "attach",
		}

		diskAttachement = append(diskAttachement, ds)
	}

	return diskAttachement
}

func newNetworkInterface(d *schema.ResourceData) oxideSDK.InstanceNetworkInterfaceAttachment {
	nis := d.Get("network_interface").([]interface{})

	if len(nis) < 1 {
		return oxideSDK.InstanceNetworkInterfaceAttachment{
			Type: "none",
		}
	}

	var interfaceCreate = []oxideSDK.NetworkInterfaceCreate{}
	for _, ni := range nis {
		nwInterface := ni.(map[string]interface{})

		nwInterfaceCreate := oxideSDK.NetworkInterfaceCreate{
			Description: nwInterface["description"].(string),
			Name:        oxideSDK.Name(nwInterface["name"].(string)),
			// TODO: Ideally from the API we should be able to create with IDs, not names
			SubnetName: oxideSDK.Name(nwInterface["subnet_name"].(string)),
			VpcName:    oxideSDK.Name(nwInterface["vpc_name"].(string)),
		}

		interfaceCreate = append(interfaceCreate, nwInterfaceCreate)
	}

	return oxideSDK.InstanceNetworkInterfaceAttachment{
		Params: interfaceCreate,
		Type:   "create",
	}
}

func networkInterfaceToState(client *oxideSDK.Client, nwInterface oxideSDK.NetworkInterfaceResultsPage) ([]interface{}, error) {
	items := nwInterface.Items
	var result = make([]interface{}, 0, len(items))
	for _, item := range items {
		var m = make(map[string]interface{})

		m["description"] = item.Description
		m["ip"] = item.Ip
		m["name"] = item.Name
		m["subnet_id"] = item.SubnetId
		m["vpc_id"] = item.VpcId

		// Ideally the NetworkInterface struct would contain the names of the VPC and subnet.
		// For now they only give the ID so we'll retrieve the names separately.
		params := oxideSDK.VpcViewParams{Vpc: oxideSDK.NameOrId(item.VpcId)}
		vpcResp, err := client.VpcView(params)
		if err != nil {
			return nil, err
		}

		subnetParams := oxideSDK.VpcSubnetViewParams{Subnet: oxideSDK.NameOrId(item.SubnetId)}
		subnetResp, err := client.VpcSubnetView(subnetParams)
		if err != nil {
			return nil, err
		}

		m["subnet_name"] = subnetResp.Name
		m["vpc_name"] = vpcResp.Name

		result = append(result, m)
	}

	return result, nil
}
