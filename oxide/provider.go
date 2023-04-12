// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

var _ provider.Provider = (*OxideProvider)(nil)

type OxideProvider struct {
	// TODO: Necesito esta madre?
	//client *oxideSDK.Client
	//	version string
}

type OxideProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

func New() provider.Provider {
	return &OxideProvider{}
}

func (p *OxideProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oxide"
	// resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *OxideProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the root of the target server",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Token used to authenticate",
			},
		},
	}
}

// Configure prepares the Oxide client for data sources and resources.
func (p *OxideProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	host := os.Getenv("OXIDE_HOST")
	token := os.Getenv("OXIDE_TOKEN")

	var data OxideProviderModel

	// Read configuration data into model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	// Check configuration data, which should take precedence over
	// environment variable data, if found.
	if data.Token.ValueString() != "" {
		token = data.Token.ValueString()
	}
	if data.Host.ValueString() != "" {
		host = data.Host.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the OXIDE_TOKEN environment variable or provider "+
				"configuration block token attribute.",
		)
	}

	if host == "" {
		resp.Diagnostics.AddError(
			"Missing Host Configuration",
			"While configuring the provider, the host was not found in "+
				"the OXIDE_HOST environment variable or provider "+
				"configuration block host attribute.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Add provider version to the user agent?
	client, err := oxideSDK.NewClient(token, "terraform-provider-oxide", host)
	if err != nil {
		resp.Diagnostics.AddError(
			"An error occurred while initializing the client for the Oxide API",
			err.Error(),
		)
		return
	}

	//p.client = client
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *OxideProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &ProjectsDataSource{} },
		//func() datasource.DataSource { return &imagesDataSource{} },
	}
}

// Resources defines the resources implemented in the provider.
func (p *OxideProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// func() resource.Resource { return &diskResource{} },
		// func() resource.Resource { return &imageResource{} },
		// func() resource.Resource { return &instanceResource{} },
		// func() resource.Resource { return &ipPoolResource{} },
		// func() resource.Resource { return &projectResource{} },
		// func() resource.Resource { return &vpcResource{} },
	}
}

//// Provider is the schema for the oxide terraform provider
//func Provider() *schema.Provider {
//	return &schema.Provider{
//		ConfigureContextFunc: newProviderMeta,
//		Schema: map[string]*schema.Schema{
//			"host": {
//				Description: "URL of the root of the target server",
//				Type:        schema.TypeString,
//				Optional:    true,
//				// We'll remove this validation for now. Will confirm later if needed with rack testing.
//				// ValidateFunc: validation.IsURLWithScheme([]string{"http", "https"}),
//				DefaultFunc: schema.MultiEnvDefaultFunc(
//					[]string{"OXIDE_HOST", "OXIDE_TEST_HOST"}, "",
//				),
//			},
//			"token": {
//				Description: "Token used to authenticate",
//				Type:        schema.TypeString,
//				Optional:    true,
//				Sensitive:   true,
//				DefaultFunc: schema.MultiEnvDefaultFunc(
//					// TODO: Decide on these tokens
//					[]string{"OXIDE_TOKEN", "OXIDE_TEST_TOKEN"}, "",
//				),
//			},
//		},
//		ResourcesMap: map[string]*schema.Resource{
//			"oxide_disk":     diskResource(),
//			"oxide_image":    imageResource(),
//			"oxide_instance": instanceResource(),
//			"oxide_ip_pool":  ipPoolResource(),
//			"oxide_project":  projectResource(),
//			"oxide_vpc":      vpcResource(),
//		},
//		DataSourcesMap: map[string]*schema.Resource{
//			"oxide_projects": projectsDataSource(),
//			"oxide_images":   imagesDataSource(),
//		},
//	}
//}

//func newProviderMeta(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
//	host := d.Get("host").(string)
//	if host == "" {
//		return nil, diag.FromErr(fmt.Errorf("host must not be empty"))
//	}
//
//	token := d.Get("token").(string)
//	if token == "" {
//		return nil, diag.FromErr(fmt.Errorf("token must not be empty"))
//	}

//	client, err := oxideSDK.NewClient(token, "terraform-provider-oxide", host)
//	if err != nil {
//		return nil, diag.FromErr(err)
//	}
//
//return client, nil
//}
