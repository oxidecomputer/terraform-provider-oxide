// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package switchportsettings

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/oxidecomputer/oxide.go/oxide"
	"github.com/stretchr/testify/require"
)

// nullTimeouts builds a null [timeouts.Value] matching the schema's timeouts
// attribute type so a model can be encoded into a [tfsdk.State].
func nullTimeouts(
	ctx context.Context,
	t *testing.T,
	s schema.Schema,
) timeouts.Value {
	t.Helper()

	attrType := s.Attributes["timeouts"].GetType()
	v, err := attrType.ValueFromTerraform(
		ctx,
		tftypes.NewValue(attrType.TerraformType(ctx), nil),
	)
	require.NoError(t, err)

	return v.(timeouts.Value)
}

// TestModelV0Upgrade verifies that switch port settings state written by a
// schema version 0 provider, which modeled BGP peers with `address` and
// `interface_name`, is migrated into the version 1 `addr` object. It also
// round-trips both models through their schemas to ensure each model stays
// consistent with the schema the framework decodes it against.
func TestModelV0Upgrade(t *testing.T) {
	ctx := context.Background()
	r := &Resource{}

	schemaV0 := r.schemaV0(ctx)

	var currentSchemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &currentSchemaResp)

	// Build a representative schema version 0 state with one numbered peer
	// (address set) and one unnumbered peer (address null).
	allowExport := &BGPPeerPeerAllowedExportResourceModel{
		Type: types.StringValue(string(oxide.ImportExportPolicyTypeNoFiltering)),
	}
	allowImport := &BGPPeerPeerAllowedImportResourceModel{
		Type: types.StringValue(string(oxide.ImportExportPolicyTypeNoFiltering)),
	}

	prior := modelV0{
		ID:          types.StringValue("3fa85f64-5717-4562-b3fc-2c963f66afa6"),
		Name:        types.StringValue("test"),
		Description: types.StringValue("a test switch port settings"),
		Addresses:   []AddressResourceModel{},
		Links:       []LinkResourceModel{},
		Routes:      []RouteResourceModel{},
		BGPPeers: []bgpPeerModelV0{
			{
				LinkName: types.StringValue("phy0"),
				Peers: []bgpPeerPeerModelV0{
					{
						Address:                iptypes.NewIPAddressValue("192.168.1.1"),
						BGPConfig:              types.StringValue("bgp-config"),
						Communities:            []types.Int64{},
						ConnectRetry:           types.Int64Value(10),
						DelayOpen:              types.Int64Value(10),
						EnforceFirstAs:         types.BoolValue(false),
						HoldTime:               types.Int64Value(10),
						IdleHoldTime:           types.Int64Value(10),
						InterfaceName:          types.StringValue("phy0"),
						Keepalive:              types.Int64Value(10),
						LocalPref:              types.Int64Null(),
						MD5AuthKey:             types.StringNull(),
						MinTTL:                 types.Int32Null(),
						MultiExitDiscriminator: types.Int64Null(),
						RemoteASN:              types.Int64Null(),
						VlanID:                 types.Int32Null(),
						AllowedExport:          allowExport,
						AllowedImport:          allowImport,
					},
					{
						Address:                iptypes.NewIPAddressNull(),
						BGPConfig:              types.StringValue("bgp-config"),
						Communities:            []types.Int64{},
						ConnectRetry:           types.Int64Value(10),
						DelayOpen:              types.Int64Value(10),
						EnforceFirstAs:         types.BoolValue(false),
						HoldTime:               types.Int64Value(10),
						IdleHoldTime:           types.Int64Value(10),
						InterfaceName:          types.StringValue("phy1"),
						Keepalive:              types.Int64Value(10),
						LocalPref:              types.Int64Null(),
						MD5AuthKey:             types.StringNull(),
						MinTTL:                 types.Int32Null(),
						MultiExitDiscriminator: types.Int64Null(),
						RemoteASN:              types.Int64Null(),
						VlanID:                 types.Int32Null(),
						AllowedExport:          allowExport,
						AllowedImport:          allowImport,
					},
				},
			},
		},
		Timeouts: nullTimeouts(ctx, t, schemaV0),
	}

	// Ensure modelV0 is consistent with schemaV0, the schema the framework
	// decodes prior state against during the upgrade.
	priorState := tfsdk.State{Schema: schemaV0}
	require.False(t, priorState.Set(ctx, prior).HasError())

	upgraded := prior.upgrade()

	require.Equal(t, "3fa85f64-5717-4562-b3fc-2c963f66afa6", upgraded.ID.ValueString())
	require.Len(t, upgraded.BGPPeers, 1)
	require.Equal(t, "phy0", upgraded.BGPPeers[0].LinkName.ValueString())
	require.Len(t, upgraded.BGPPeers[0].Peers, 2)

	numbered := upgraded.BGPPeers[0].Peers[0]
	require.NotNil(t, numbered.Addr)
	require.Equal(t, string(oxide.RouterPeerTypeTypeNumbered), numbered.Addr.Type.ValueString())
	require.Equal(t, "192.168.1.1", numbered.Addr.IP.ValueString())
	require.True(t, numbered.Addr.RouterLifetime.IsNull())

	unnumbered := upgraded.BGPPeers[0].Peers[1]
	require.NotNil(t, unnumbered.Addr)
	require.Equal(t, string(oxide.RouterPeerTypeTypeUnnumbered), unnumbered.Addr.Type.ValueString())
	require.True(t, unnumbered.Addr.IP.IsNull())
	require.True(t, unnumbered.Addr.RouterLifetime.IsNull())

	// Ensure the upgraded model is consistent with the current schema, the
	// schema the framework encodes the upgraded state against.
	currentState := tfsdk.State{Schema: currentSchemaResp.Schema}
	require.False(t, currentState.Set(ctx, upgraded).HasError())
}
