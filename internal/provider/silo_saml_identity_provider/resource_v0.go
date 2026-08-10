// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package silosamlidp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// schemaV0 is the resource schema before idp_metadata_source and
// signing_keypair became write-only.
func (r *Resource) schemaV0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages a SAML identity provider (IdP) for an Oxide silo.

-> This resource does not support updates. All attributes are immutable once
created.

-> This resource does not support deletion from the Oxide API. When destroyed in
Terraform, it will be removed from state but will continue to exist in Oxide.
`,
		Attributes: map[string]schema.Attribute{
			"silo": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the silo.",
			},
			"acs_url": schema.StringAttribute{
				Required:    true,
				Description: "URL where the identity provider should send the SAML response.",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Free-form text describing the SAML identity provider.",
			},
			"group_attribute_name": schema.StringAttribute{
				Optional:    true,
				Description: "SAML attribute that holds a user's group membership.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the SAML identity provider.",
			},
			"idp_entity_id": schema.StringAttribute{
				Required:    true,
				Description: "Identity provider's entity ID.",
			},
			"idp_metadata_source": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Source of identity provider metadata (URL or base64-encoded XML).",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The type of metadata source. Must be one of: `url`, `base64_encoded_xml`.",
						Validators: []validator.String{
							stringvalidator.OneOf(
								"url",
								"base64_encoded_xml",
							),
						},
					},
					"url": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "URL to fetch metadata from (required when type is `url`). Conflicts with `data`.",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("data"),
							),
						},
					},
					"data": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Base64-encoded XML metadata (required when type is `base64_encoded_xml`). Conflicts with `url`.",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("url"),
							),
						},
					},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique, immutable, user-controlled identifier of the SAML identity provider.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(63),
				},
			},
			"signing_keypair": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "RSA private key and public certificate for signing SAML requests.",
				Attributes: map[string]schema.Attribute{
					"private_key": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "RSA private key (base64 encoded).",
					},
					"public_cert": schema.StringAttribute{
						Required:    true,
						Description: "Public certificate (base64 encoded).",
					},
				},
			},
			"slo_url": schema.StringAttribute{
				Required:    true,
				Description: "URL where the identity provider should send logout requests.",
			},
			"sp_client_id": schema.StringAttribute{
				Required:    true,
				Description: "Service provider's client ID.",
			},
			"technical_contact_email": schema.StringAttribute{
				Required:    true,
				Description: "Technical contact email for SAML configuration.",
			},
			"time_created": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this SAML identity provider was created.",
			},
			"time_modified": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when this SAML identity provider was last modified.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
			}),
		},
	}
}

// stateUpgraderV0 updates the v0 state to current.
func (r *Resource) stateUpgraderV0(
	ctx context.Context,
	req resource.UpgradeStateRequest,
	resp *resource.UpgradeStateResponse,
) {
	var state ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Explicitly remove write-only attributes from state. The Terraform plugin
	// framework does this automatically but having this here makes it easier to
	// understand the intentions of the state upgrade.
	state.IdpMetadataSource = nil
	if state.SigningKeypair != nil {
		state.SigningKeypair.PrivateKey = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
