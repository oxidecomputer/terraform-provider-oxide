// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oxidecomputer/oxide.go/oxide"
)

func TestNewMulticastGroupJoinSpecs(t *testing.T) {
	sourceIPs := mustStringSet(t, "192.0.2.1", "192.0.2.2")

	got, diags := newMulticastGroupJoinSpecs([]MulticastGroupModel{
		{
			Group:     types.StringValue("multicast-group"),
			IPVersion: types.StringValue(string(oxide.IpVersionV4)),
			SourceIPs: sourceIPs,
		},
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if len(got) != 1 {
		t.Fatalf("got %d multicast groups, want 1", len(got))
	}
	if got[0].Group != oxide.MulticastGroupIdentifier("multicast-group") {
		t.Fatalf("got group %q, want multicast-group", got[0].Group)
	}
	if got[0].IpVersion != oxide.IpVersionV4 {
		t.Fatalf("got IP version %q, want %q", got[0].IpVersion, oxide.IpVersionV4)
	}
	if len(got[0].SourceIps) != 2 {
		t.Fatalf("got %d source IPs, want 2", len(got[0].SourceIps))
	}
}

func TestMulticastGroupsChangedIgnoresSourceIPOrder(t *testing.T) {
	a := []MulticastGroupModel{
		{
			Group:     types.StringValue("multicast-group"),
			IPVersion: types.StringNull(),
			SourceIPs: mustStringSet(t, "192.0.2.1", "192.0.2.2"),
		},
	}
	b := []MulticastGroupModel{
		{
			Group:     types.StringValue("multicast-group"),
			IPVersion: types.StringNull(),
			SourceIPs: mustStringSet(t, "192.0.2.2", "192.0.2.1"),
		},
	}

	if multicastGroupsChanged(a, b) {
		t.Fatal("expected source IP order not to produce a multicast group change")
	}
}

func TestMatchingMulticastGroupModel(t *testing.T) {
	models := []MulticastGroupModel{
		{Group: types.StringValue("group-name")},
		{Group: types.StringValue("224.0.2.42")},
		{Group: types.StringValue("group-id")},
	}
	member := oxide.MulticastGroupMember{
		MulticastGroupId: "group-id",
		MulticastIp:      "224.0.2.42",
		Name:             oxide.Name("group-name"),
	}

	got := matchingMulticastGroupModel(member, models)
	if got == nil {
		t.Fatal("expected multicast group model match")
	}
	if got.Group.ValueString() != "group-name" {
		t.Fatalf("got group %q, want first configured matching identifier", got.Group.ValueString())
	}
}

func mustStringSet(t *testing.T, values ...string) types.Set {
	t.Helper()

	attrs := make([]attr.Value, 0, len(values))
	for _, value := range values {
		attrs = append(attrs, types.StringValue(value))
	}
	set, diags := types.SetValue(types.StringType, attrs)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	return set
}
