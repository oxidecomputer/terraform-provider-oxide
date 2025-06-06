// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var _ provider.Provider = (*oxideProvider)(nil)

type oxideProvider struct {
	// TODO: This variable should be updated to the non-dev version
	// during the release process. Double check.
	version string
}

type oxideProviderModel struct {
	Host    types.String `tfsdk:"host"`
	Token   types.String `tfsdk:"token"`
	Profile types.String `tfsdk:"profile"`
}

// New initialises a new provider
func New() provider.Provider {
	return &oxideProvider{
		version: Version,
	}
}

func (p *oxideProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "oxide"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *oxideProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the root of the target server",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("profile"),
					}...),
				},
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Token used to authenticate",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("profile"),
					}...),
				},
			},
			"profile": schema.StringAttribute{
				Optional:    true,
				Description: "Profile used to authenticate. Retrieves host and token from credentials.toml",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("host"),
						path.MatchRoot("token"),
					}...),
				},
			},
		},
	}
}

// Configure prepares the Oxide client for data sources and resources.
func (p *oxideProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Oxide client")

	host := os.Getenv("OXIDE_HOST")
	token := os.Getenv("OXIDE_TOKEN")
	profile := ""

	var data oxideProviderModel

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
	if data.Profile.ValueString() != "" {
		profile = data.Profile.ValueString()
	}

	if token == "" && profile == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the OXIDE_TOKEN environment variable or "+
				"configuration block token attribute, or profile.",
		)
	}

	if host == "" && profile == "" {
		resp.Diagnostics.AddError(
			"Missing Host Configuration",
			"While configuring the provider, the host was not found in "+
				"the OXIDE_HOST environment variable or "+
				"configuration block host attribute, or profile.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "host", host)
	ctx = tflog.SetField(ctx, "token", token)
	ctx = tflog.SetField(ctx, "profile", profile)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "token")
	tflog.Debug(ctx, "Creating Oxide client")

	config := oxide.Config{
		Token:     token,
		UserAgent: fmt.Sprintf("terraform-provider-oxide/%s", Version),
		Host:      host,
		Profile:   profile,
	}
	client, err := oxide.NewClient(&config)
	if err != nil {
		resp.Diagnostics.AddError(
			"An error occurred while initializing the client for the Oxide API",
			err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Configured Oxide client", map[string]any{"success": true})

	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *oxideProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAntiAffinityGroupDataSource,
		NewImageDataSource,
		NewImagesDataSource,
		NewInstanceExternalIPsDataSource,
		NewIpPoolDataSource,
		NewProjectDataSource,
		NewProjectsDataSource,
		NewSiloDataSource,
		NewSSHKeyDataSource,
		NewVPCDataSource,
		NewVPCInternetGatewayDataSource,
		NewVPCRouterDataSource,
		NewVPCRouterRouteDataSource,
		NewVPCSubnetDataSource,
		NewFloatingIPDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *oxideProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAntiAffinityGroupResource,
		NewDiskResource,
		NewImageResource,
		NewInstanceResource,
		NewIPPoolResource,
		NewIpPoolSiloLinkResource,
		NewProjectResource,
		NewSnapshotResource,
		NewSSHKeyResource,
		NewVPCResource,
		NewVPCInternetGatewayResource,
		NewVPCFirewallRulesResource,
		NewVPCRouterResource,
		NewVPCRouterRouteResource,
		NewVPCSubnetResource,
		NewFloatingIPResource,
		NewSiloResource,
	}
}
