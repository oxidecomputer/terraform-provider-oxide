// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
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
var _ provider.ProviderWithFunctions = (*oxideProvider)(nil)

type oxideProvider struct {
	// TODO: This variable should be updated to the non-dev version
	// during the release process. Double check.
	version string
}

type oxideProviderModel struct {
	Host               types.String `tfsdk:"host"`
	Token              types.String `tfsdk:"token"`
	Profile            types.String `tfsdk:"profile"`
	ConfigDir          types.String `tfsdk:"config_dir"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

// New initialises a new provider
func New() provider.Provider {
	return &oxideProvider{
		version: Version,
	}
}

func (p *oxideProvider) Metadata(
	ctx context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "oxide"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *oxideProvider) Schema(
	_ context.Context,
	_ provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The Oxide provider is used to declaratively manage
[Oxide](https://oxide.computer) infrastructure.

The provider uses the [Oxide Go SDK](https://github.com/oxidecomputer/oxide.go)
to create, read, update, and delete Oxide resources.
`,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "Oxide API host (e.g., https://oxide.sys.example.com). Conflicts with `profile`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("profile"),
					}...),
				},
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Oxide API token. Conflicts with `profile`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("profile"),
					}...),
				},
			},
			"profile": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Profile to load from the Oxide credentials file. Conflicts with `host` and `token`.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("host"),
						path.MatchRoot("token"),
					}...),
				},
			},
			"config_dir": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The directory to search for Oxide credentials file.",
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Disables TLS certificate if `true`. This is insecure and should only be used for testing or in controlled environments.",
			},
		},
	}
}

// Configure prepares the Oxide client for data sources and resources.
func (p *oxideProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	tflog.Info(ctx, "Configuring Oxide client")

	var data oxideProviderModel

	// Read configuration data into model.
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	clientOpts := []oxide.ClientOption{
		oxide.WithUserAgent(fmt.Sprintf("terraform-provider-oxide/%s", Version)),
	}

	if host := data.Host.ValueString(); host != "" {
		clientOpts = append(clientOpts, oxide.WithHost(host))
	}
	if token := data.Token.ValueString(); token != "" {
		clientOpts = append(clientOpts, oxide.WithToken(token))
	}
	if profile := data.Profile.ValueString(); profile != "" {
		clientOpts = append(clientOpts, oxide.WithProfile(profile))
	}
	if dir := data.ConfigDir.ValueString(); dir != "" {
		clientOpts = append(clientOpts, oxide.WithConfigDir(dir))
	}
	if data.InsecureSkipVerify.ValueBool() {
		clientOpts = append(clientOpts, oxide.WithInsecureSkipVerify())
	}

	client, err := oxide.NewClient(clientOpts...)
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
		NewAddressLotDataSource,
		NewAntiAffinityGroupDataSource,
		NewDiskDataSource,
		NewFloatingIPDataSource,
		NewImageDataSource,
		NewImagesDataSource,
		NewInstanceExternalIPsDataSource,
		NewIpPoolDataSource,
		NewProjectDataSource,
		NewProjectsDataSource,
		NewSiloDataSource,
		NewSSHKeyDataSource,
		NewSubnetPoolDataSource,
		NewSystemIpPoolsDataSource,
		NewVPCDataSource,
		NewVPCInternetGatewayDataSource,
		NewVPCRouterDataSource,
		NewVPCRouterRouteDataSource,
		NewVPCSubnetDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *oxideProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAddressLotResource,
		NewAntiAffinityGroupResource,
		NewDiskResource,
		NewExternalSubnetResource,
		NewFloatingIPResource,
		NewImageResource,
		NewInstanceResource,
		NewIPPoolResource,
		NewIpPoolSiloLinkResource,
		NewProjectResource,
		NewSiloResource,
		NewSiloSamlIdentityProviderResource,
		NewSnapshotResource,
		NewSSHKeyResource,
		NewSubnetPoolMemberResource,
		NewSubnetPoolResource,
		NewSubnetPoolSiloLinkResource,
		NewSwitchPortSettingsResource,
		NewVPCFirewallRulesResource,
		NewVPCInternetGatewayResource,
		NewVPCResource,
		NewVPCRouterResource,
		NewVPCRouterRouteResource,
		NewVPCSubnetResource,
	}
}

// Functions defines the functions implemented in the provider.
func (p *oxideProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		NewToVPCFirewallRulesMapFunction,
	}
}
