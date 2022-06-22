// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go"
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
		"project_id": {
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

	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	hostName := d.Get("host_name").(string)
	memory := d.Get("memory").(int)
	ncpus := d.Get("ncpus").(int)

	body := oxideSDK.InstanceCreate{
		Description:       description,
		Name:              name,
		Hostname:          hostName,
		Memory:            oxideSDK.ByteCount(memory),
		NCPUs:             oxideSDK.InstanceCPUCount(ncpus),
		Disks:             newInstanceDiskAttach(d),
		NetworkInterfaces: newNetworkInterface(d),
	}

	resp, err := client.Instances.Create(orgName, projectName, &body)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.ID)

	return readInstance(ctx, d, meta)
}

func readInstance(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	instanceName := d.Get("name").(string)
	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)

	resp, err := client.Instances.Get(instanceName, orgName, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := instanceToState(d, resp); err != nil {
		return diag.FromErr(err)
	}

	resp2, err := client.Instances.NetworkInterfacesList(1000000, "", oxideSDK.NameSortModeNameAscending, instanceName, orgName, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	networkInterfaces := networkInterfaceToState(*resp2)
	if err := d.Set("network_interface", networkInterfaces); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateInstance(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TODO: Currently there is no endpoint to update an instance. This function will remain
	// as readonly until such endpoint exists.
	return readInstance(ctx, d, meta)
}

func deleteInstance(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)
	instanceName := d.Get("name").(string)
	orgName := d.Get("organization_name").(string)
	projectName := d.Get("project_name").(string)

	_, err := client.Instances.Stop(instanceName, orgName, projectName)
	if err != nil {
		if is404(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Wait for instance to be stopped before attempting to destroy
	ch := make(chan error)
	go waitForStoppedInstance(client, instanceName, orgName, projectName, ch)
	e := <-ch
	if e != nil {
		return diag.FromErr(e)
	}

	if err := client.Instances.Delete(instanceName, orgName, projectName); err != nil {
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
	if err := d.Set("id", instance.ID); err != nil {
		return err
	}
	if err := d.Set("project_id", instance.ProjectID); err != nil {
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

func waitForStoppedInstance(client *oxideSDK.Client, instanceName, orgName, projectName string, ch chan error) {
	for {
		resp, err := client.Instances.Get(instanceName, orgName, projectName)
		if err != nil {
			ch <- err
		}
		if resp.RunState == "stopped" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	ch <- nil
}

func newInstanceDiskAttach(d *schema.ResourceData) []oxideSDK.InstanceDiskAttachment {
	var diskAttachement = []oxideSDK.InstanceDiskAttachment{}
	disks := d.Get("attach_to_disks").([]interface{})

	if len(disks) < 1 {
		return diskAttachement
	}
	for _, disk := range disks {
		ds := oxideSDK.InstanceDiskAttachment{
			Name: disk.(string),
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
			Name:        nwInterface["name"].(string),
			SubnetName:  nwInterface["subnet_name"].(string),
			VPCName:     nwInterface["vpc_name"].(string),
		}

		interfaceCreate = append(interfaceCreate, nwInterfaceCreate)
	}

	return oxideSDK.InstanceNetworkInterfaceAttachment{
		Params: interfaceCreate,
		Type:   "create",
	}
}

func networkInterfaceToState(nwInterface oxideSDK.NetworkInterfaceResultsPage) []interface{} {
	items := nwInterface.Items
	var result = make([]interface{}, 0, len(items))
	for _, item := range items {
		var m = make(map[string]interface{})

		m["description"] = item.Description
		m["ip"] = item.Ip
		m["name"] = item.Name
		m["subnet_id"] = item.SubnetID
		m["vpc_id"] = item.VPCId

		// TODO: Unfortunately NetworkInterface doesn't have the following fields yet.
		// This means that they are unset when a read is performed (which is on every create).
		// For the demo this won't be a problem, but we must fix this before we can implement
		// update.
		//m["subnet_name"] = item.SubnetName
		//m["vpc_name"] = item.VPCName

		result = append(result, m)
	}

	return result
}
