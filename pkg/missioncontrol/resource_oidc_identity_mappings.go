package missioncontrol

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
)

const (
	IdentityMappingsEndpoint = "access/api/v1/oidc/{provider_name}/identity_mappings/{mapping_name}"
	IdentityMappingsEndpoint  = "access/api/v1/oidc/{provider_name}/identity_mappings/{mapping_name}"
)

func NewOidcIdentityMappingsResource() resource.Resource {
	return &OidcIdentityMappingsResource{
		TypeName: "mission-control_identity_mappings",
	}
}

type OidcIdentityMappingsResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type OidcIdentityMappingsResourceModel struct {
	ProviderName types.String `tfsdk:"provider_name"`
	MappingName types.String `tfsdk:"mapping_name"`
}

type IdentityMappingsRequestAPIModel struct {
	ProviderName string `json:"provider_name"`
	MappingName string `json:"mapping_name"`
}

type IdentityMappingsAPIModel struct {
	ProviderName string `json:"provider_name"`
	MappingName string `json:"mapping_name"`
}

func (r *OidcIdentityMappingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *OidcIdentityMappingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"provider_name": schema.StringAttribute{
				Required: true,
				Description: "The provider_name of the resource.",
			},
			"mapping_name": schema.StringAttribute{
				Required: true,
				Description: "The mapping_name of the resource.",
			},
		},
		MarkdownDescription: "Manages identity_mappings in JFrog Mission-Control.",
	}
}

func (r *OidcIdentityMappingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}


func (r *OidcIdentityMappingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan OidcIdentityMappingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := OidcIdentityMappingsRequestAPIModel{
		ProviderName: plan.ProviderName.ValueString(),
		MappingName: plan.MappingName.ValueString(),
	}

	var result OidcIdentityMappingsAPIModel

	response, err := r.ProviderData.Client.R().
		SetBody(requestBody).
		SetResult(&result).
		Post(IdentityMappingsEndpoint)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}


func (r *OidcIdentityMappingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state OidcIdentityMappingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result OidcIdentityMappingsAPIModel

	response, err := r.ProviderData.Client.R().
		SetPathParams(map[string]string{
			"provider_name": state.ProviderName.ValueString(),
			"mapping_name": state.MappingName.ValueString(),
		}).
		SetResult(&result).
		Get(IdentityMappingsEndpoint)
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}


func (r *OidcIdentityMappingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan OidcIdentityMappingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	requestBody := OidcIdentityMappingsRequestAPIModel{
		ProviderName: plan.ProviderName.ValueString(),
		MappingName: plan.MappingName.ValueString(),
	}

	response, err := r.ProviderData.Client.R().
		SetPathParams(map[string]string{
			"provider_name": plan.ProviderName.ValueString(),
			"mapping_name": plan.MappingName.ValueString(),
		}).
		SetBody(requestBody).
		Put(IdentityMappingsEndpoint)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}



func (r *OidcIdentityMappingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state OidcIdentityMappingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	response, err := r.ProviderData.Client.R().
		SetPathParams(map[string]string{
			"provider_name": state.ProviderName.ValueString(),
			"mapping_name": state.MappingName.ValueString(),
		}).
		Delete(IdentityMappingsEndpoint)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		return
	}

	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}
}

