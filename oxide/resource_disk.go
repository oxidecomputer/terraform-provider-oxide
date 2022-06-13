package oxide

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go"
)

func diskResource() *schema.Resource {
	return &schema.Resource{
		Description:   "",
		Schema:        newDiskSchema(),
		CreateContext: createDisk,
		ReadContext:   readDisk,
		UpdateContext: updateDisk,
		DeleteContext: deleteDisk,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func newDiskSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{}
}

func createDisk(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*oxideSDK.Client)

	organizationName := "maze-war"
	projectName := "prod-online"
	body := oxideSDK.DiskCreate{
		Description: "",
		Name:        "cio-api",
		DiskSource: oxideSDK.DiskSource{
			BlockSize: oxideSDK.BlockSize(512),
			Type:      "blank",
		},
		Size: 1024,
	}

	resp, err := client.Disks.Create(organizationName, projectName, &body)
	if err != nil {
		panic(err)
	}

	d.SetId(resp.ProjectID)

	return readDisk(ctx, d, meta)
}

func readDisk(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func updateDisk(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func deleteDisk(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
