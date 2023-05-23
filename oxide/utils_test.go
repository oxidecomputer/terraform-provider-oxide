// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_isIPv4(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "success",
			args: args{"172.20.15.227"},
			want: true,
		},
		{
			name: "fail with IPv6",
			args: args{"2001:0db8:3c4d:0015:0000:0000:1a2f:1a2b"},
			want: false,
		},
		{
			name: "fail with random input",
			args: args{"totally-legit-ipv4"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isIPv4(tt.args.str)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_isIPv6(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "success",
			args: args{"2001:0db8:3c4d:0015:0000:0000:1a2f:1a2b"},
			want: true,
		},
		{
			name: "fail with IPv4",
			args: args{"172.20.15.227"},
			want: false,
		},
		{
			name: "fail with random input",
			args: args{"totally-legit-ipv6"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isIPv6(tt.args.str)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_difference(t *testing.T) {
	type args struct {
		a []attr.Value
		b []attr.Value
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "success",
			args: args{
				a: []attr.Value{
					types.StringValue("one"),
					types.StringValue("two"),
					types.StringValue("three"),
				},
				b: []attr.Value{
					types.StringValue("one"),
				},
			},
			want: []string{"\"two\"", "\"three\""},
		},
		{
			name: "retrieves multiple items if there are duplicate entries",
			args: args{
				a: []attr.Value{
					types.StringValue("one"),
					types.StringValue("two"),
					types.StringValue("two"),
					types.StringValue("three"),
				},
				b: []attr.Value{
					types.StringValue("one"),
				},
			},
			want: []string{"\"two\"", "\"two\"", "\"three\""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := difference(tt.args.a, tt.args.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
