// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var (
	_ resource.Resource              = (*siloSamlIdentityProvider)(nil)
	_ resource.ResourceWithConfigure = (*siloSamlIdentityProvider)(nil)
)

// NewSiloSamlIdentityProviderResource returns a new silo SAML identity provider resource.
func NewSiloSamlIdentityProviderResource() resource.Resource {
	return &siloSamlIdentityProvider{}
}

type siloSamlIdentityProvider struct {
	client *oxide.Client
}

type siloSamlIdentityProviderResourceModel struct {
	ID                    types.String                                 `tfsdk:"id"`
	Name                  types.String                                 `tfsdk:"name"`
	Description           types.String                                 `tfsdk:"description"`
	Silo                  types.String                                 `tfsdk:"silo"`
	AcsUrl                types.String                                 `tfsdk:"acs_url"`
	IdpEntityId           types.String                                 `tfsdk:"idp_entity_id"`
	SloUrl                types.String                                 `tfsdk:"slo_url"`
	SpClientId            types.String                                 `tfsdk:"sp_client_id"`
	TechnicalContactEmail types.String                                 `tfsdk:"technical_contact_email"`
	IdpMetadataSource     *siloSamlIdentityProviderMetadataSourceModel `tfsdk:"idp_metadata_source"`
	GroupAttributeName    types.String                                 `tfsdk:"group_attribute_name"`
	SigningKeypair        *siloSamlIdentityProviderSigningKeypairModel `tfsdk:"signing_keypair"`
	TimeCreated           types.String                                 `tfsdk:"time_created"`
	TimeModified          types.String                                 `tfsdk:"time_modified"`
	Timeouts              timeouts.Value                               `tfsdk:"timeouts"`
}

type siloSamlIdentityProviderSigningKeypairModel struct {
	PrivateKey types.String `tfsdk:"private_key"`
	PublicCert types.String `tfsdk:"public_cert"`
}

type siloSamlIdentityProviderMetadataSourceModel struct {
	Type types.String `tfsdk:"type"`
	Url  types.String `tfsdk:"url"`
	Data types.String `tfsdk:"data"`
}

func (r *siloSamlIdentityProvider) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = "oxide_silo_saml_identity_provider"
}

func (r *siloSamlIdentityProvider) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *siloSamlIdentityProvider) Schema(
	ctx context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
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

func (r *siloSamlIdentityProvider) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan siloSamlIdentityProviderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	var idpMetadataSource oxide.IdpMetadataSource
	switch oxide.IdpMetadataSourceType(plan.IdpMetadataSource.Type.ValueString()) {
	case oxide.IdpMetadataSourceTypeBase64EncodedXml:
		idpMetadataSource = oxide.IdpMetadataSource{
			Value: &oxide.IdpMetadataSourceBase64EncodedXml{
				Data: plan.IdpMetadataSource.Data.ValueString(),
			},
		}
	case oxide.IdpMetadataSourceTypeUrl:
		idpMetadataSource = oxide.IdpMetadataSource{
			Value: &oxide.IdpMetadataSourceUrl{
				Url: plan.IdpMetadataSource.Url.ValueString(),
			},
		}
	}

	params := oxide.SamlIdentityProviderCreateParams{
		Silo: oxide.NameOrId(plan.Silo.ValueString()),
		Body: &oxide.SamlIdentityProviderCreate{
			AcsUrl:                plan.AcsUrl.ValueString(),
			IdpEntityId:           plan.IdpEntityId.ValueString(),
			Name:                  oxide.Name(plan.Name.ValueString()),
			SloUrl:                plan.SloUrl.ValueString(),
			SpClientId:            plan.SpClientId.ValueString(),
			TechnicalContactEmail: plan.TechnicalContactEmail.ValueString(),
			IdpMetadataSource:     idpMetadataSource,
			Description:           plan.Description.ValueString(),
			GroupAttributeName:    plan.GroupAttributeName.ValueString(),
		},
	}

	if plan.SigningKeypair != nil {
		params.Body.SigningKeypair = oxide.DerEncodedKeyPair{
			PrivateKey: plan.SigningKeypair.PrivateKey.ValueString(),
			PublicCert: plan.SigningKeypair.PublicCert.ValueString(),
		}
	}

	idpConfig, err := r.client.SamlIdentityProviderCreate(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating SAML identity provider",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("created SAML identity provider with ID: %v", idpConfig.Id),
		map[string]any{"success": true},
	)

	plan.ID = types.StringValue(idpConfig.Id)
	plan.TimeCreated = types.StringValue(idpConfig.TimeCreated.String())
	plan.TimeModified = types.StringValue(idpConfig.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *siloSamlIdentityProvider) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state siloSamlIdentityProviderResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, defaultTimeout())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	params := oxide.SamlIdentityProviderViewParams{
		Provider: oxide.NameOrId(state.ID.ValueString()),
	}

	idpConfig, err := r.client.SamlIdentityProviderView(ctx, params)
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Unable to read SAML identity provider:",
			"API error: "+err.Error(),
		)
		return
	}

	tflog.Trace(
		ctx,
		fmt.Sprintf("read SAML identity provider with ID: %v", idpConfig.Id),
		map[string]any{"success": true},
	)

	state.ID = types.StringValue(idpConfig.Id)
	state.TimeCreated = types.StringValue(idpConfig.TimeCreated.String())
	state.TimeModified = types.StringValue(idpConfig.TimeModified.String())
	state.AcsUrl = types.StringValue(idpConfig.AcsUrl)
	state.Description = types.StringValue(idpConfig.Description)
	state.GroupAttributeName = types.StringValue(idpConfig.GroupAttributeName)
	state.IdpEntityId = types.StringValue(idpConfig.IdpEntityId)
	state.Name = types.StringValue(string(idpConfig.Name))
	state.SloUrl = types.StringValue(idpConfig.SloUrl)
	state.SpClientId = types.StringValue(idpConfig.SpClientId)
	state.TechnicalContactEmail = types.StringValue(idpConfig.TechnicalContactEmail)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update is purposefully unsupported since there's no upstream API to update a
// silo's SAML identity provider configuration.
//
// An error is added to the diagnostics so that Terraform will stop execution
// and prompt the user to fix their configuration.
func (r *siloSamlIdentityProvider) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	resp.Diagnostics.AddError(
		"The oxide_silo_saml_identity_provider resource does not support updates.",
		"This resource represents immutable silo SAML identity provider configuration. Please update your Terraform configuration to match the state.",
	)
}

// Delete is purposefully unsupported since there's no upstream API to delete a
// silo's SAML identity provider configuration.
//
// A warning is added to the diagnostics so that the resource will be removed
// from the state and Terraform execution will continue and prompt the user on
// what happened.
func (r *siloSamlIdentityProvider) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	resp.Diagnostics.AddWarning(
		"The oxide_silo_saml_identity_provider resource does not support deletion.",
		"This resource represents immutable silo SAML identity provider configuration. The resource will be removed from Terraform state but not from Oxide.",
	)
}
