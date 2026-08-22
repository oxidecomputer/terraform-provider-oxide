// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippool

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
)

func TestResourceSchemaAndMetadata(t *testing.T) {
	t.Parallel()

	res := NewResource()
	metadataResponse := &resource.MetadataResponse{}
	res.Metadata(context.Background(), resource.MetadataRequest{}, metadataResponse)
	require.Equal(t, "oxide_system_ip_pool", metadataResponse.TypeName)

	schemaResponse := &resource.SchemaResponse{}
	res.Schema(context.Background(), resource.SchemaRequest{}, schemaResponse)
	require.Empty(t, schemaResponse.Schema.DeprecationMessage)
	require.NotContains(t, schemaResponse.Schema.Attributes, "ranges")

	description, ok := schemaResponse.Schema.Attributes["description"].(schema.StringAttribute)
	require.True(t, ok)
	require.False(t, description.Required)
	require.True(t, description.Optional)
	require.True(t, description.Computed)
	require.Nil(t, description.Default)
}
