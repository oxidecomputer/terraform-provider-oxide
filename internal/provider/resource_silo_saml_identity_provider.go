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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/oxidecomputer/oxide.go/oxide"
)

var (
	_ resource.Resource                = (*siloSamlIdentityProvider)(nil)
	_ resource.ResourceWithConfigure   = (*siloSamlIdentityProvider)(nil)
	_ resource.ResourceWithImportState = (*siloSamlIdentityProvider)(nil)
)

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

func (r *siloSamlIdentityProvider) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "oxide_silo_saml_identity_provider"
}

func (r *siloSamlIdentityProvider) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*oxide.Client)
}

func (r *siloSamlIdentityProvider) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *siloSamlIdentityProvider) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"silo": schema.StringAttribute{
				Required:    true,
				Description: "Name or ID of the silo.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"acs_url": schema.StringAttribute{
				Required:    true,
				Description: "URL where the IdP should send the SAML response.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Free-form text describing this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_attribute_name": schema.StringAttribute{
				Optional:    true,
				Description: "SAML attribute that holds a user's group membership.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique, immutable, system-controlled identifier of the SAML identity provider.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"idp_entity_id": schema.StringAttribute{
				Required:    true,
				Description: "Identity provider's entity ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"idp_metadata_source": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Source of identity provider metadata (URL or base64-encoded XML).",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"url",
								"base64_encoded_xml",
							),
						},
					},
					"url": schema.StringAttribute{
						Optional:    true,
						Description: "URL to fetch metadata from (required when type is 'url').",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("data")),
						},
					},
					"data": schema.StringAttribute{
						Optional:    true,
						Description: "Base64-encoded XML metadata (required when type is 'base64_encoded_xml').",
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("url")),
						},
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique, immutable, user-controlled identifier of the SAML identity provider.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(63),
				},
			},
			"signing_keypair": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "RSA private key and public certificate for signing.",
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
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"slo_url": schema.StringAttribute{
				Required:    true,
				Description: "URL where the IdP should send logout requests.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sp_client_id": schema.StringAttribute{
				Required:    true,
				Description: "Service provider's client ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"technical_contact_email": schema.StringAttribute{
				Required:    true,
				Description: "Technical contact email for SAML configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *siloSamlIdentityProvider) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

	idpMetadataSource := oxide.IdpMetadataSource{
		Type: oxide.IdpMetadataSourceType(plan.IdpMetadataSource.Type.ValueString()),
	}

	switch idpMetadataSource.Type {
	case oxide.IdpMetadataSourceTypeBase64EncodedXml:
		idpMetadataSource.Data = plan.IdpMetadataSource.Data.ValueString()
	case oxide.IdpMetadataSourceTypeUrl:
		idpMetadataSource.Url = plan.IdpMetadataSource.Url.ValueString()
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

	tflog.Trace(ctx, fmt.Sprintf("created SAML identity provider with ID: %v", idpConfig.Id), map[string]any{"success": true})

	plan.ID = types.StringValue(idpConfig.Id)
	plan.TimeCreated = types.StringValue(idpConfig.TimeCreated.String())
	plan.TimeModified = types.StringValue(idpConfig.TimeModified.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *siloSamlIdentityProvider) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

	tflog.Trace(ctx, fmt.Sprintf("read SAML identity provider with ID: %v", idpConfig.Id), map[string]any{"success": true})

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

func (r *siloSamlIdentityProvider) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The oxide_silo_idp_configuration resource does not support updates. "+
			"Changes require replacement of the resource.",
	)
}

func (r *siloSamlIdentityProvider) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Delete Not Supported",
		"The oxide_silo_idp_configuration resource does not support deletion. "+
			"This resource represents immutable SAML identity provider configuration.",
	)
}
