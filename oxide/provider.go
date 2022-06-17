// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	oxideSDK "github.com/oxidecomputer/oxide.go"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ConfigureContextFunc: newProviderMeta,
		Schema: map[string]*schema.Schema{
			"host": {
				Description:  "URL of the root of the target server",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsURLWithScheme([]string{"http", "https"}),
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{"OXIDE_HOST", "OXIDE_TEST_HOST"}, "",
				),
			},
			"token": {
				Description: "Token used to authenticate",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.MultiEnvDefaultFunc(
					// TODO: Decide on these tokens
					[]string{"OXIDE_TOKEN", "OXIDE_TEST_TOKEN"}, "",
				),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"oxide_disk":     diskResource(),
			"oxide_instance": instanceResource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"oxide_organizations": organizationsDataSource(),
			"oxide_projects":      projectsDataSource(),
			"oxide_global_images": globalImagesDataSource(),
		},
	}
}

func newProviderMeta(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	if host == "" {
		return nil, diag.FromErr(fmt.Errorf("host must not be empty"))
	}

	token := d.Get("token").(string)
	if token == "" {
		return nil, diag.FromErr(fmt.Errorf("token must not be empty"))
	}

	client, err := oxideSDK.NewClient(token, "terraform-provider-oxide", host)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return client, nil
}
