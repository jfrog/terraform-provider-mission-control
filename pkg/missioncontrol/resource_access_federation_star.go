package missioncontrol

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validator_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const accessFederationEndpoint = "mc/api/v1/federation/{id}"

var _ resource.Resource = &accessFederationStarResource{}

type accessFederationStarResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

func NewAccessFederationStarResource() resource.Resource {
	return &accessFederationStarResource{
		TypeName: "missioncontrol_access_federation_star",
	}
}

func (r *accessFederationStarResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *accessFederationStarResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "ID for the source Platform Deployment. Use [Get Access Federation Candidate API](https://jfrog.com/help/r/jfrog-rest-apis/get-access-federation-candidates) to get a list of ID.",
			},
			"entities": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf("USERS", "GROUPS", "PERMISSIONS", "TOKENS"),
					),
				},
				Description: "Entity types to sync. Allow values: `USERS`, `GROUPS`, `PERMISSIONS`, `TOKENS`",
			},
			"targets": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
							Description: "ID of the targeted Platform Deployment",
						},
						"url": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								validator_string.IsURLHttpOrHttps(),
								stringvalidator.RegexMatches(regexp.MustCompile(`^.+/access$`), "must end in '/access'"),
							},
							Description: "Target Platform deployment URL: http://<hostname>:<port>/access; for example: http://myplatformserver:8082/access.",
						},
						"permission_filters": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"include_patterns": schema.SetAttribute{
									ElementType: types.StringType,
									Optional:    true,
								},
								"exclude_patterns": schema.SetAttribute{
									ElementType: types.StringType,
									Optional:    true,
								},
							},
							Optional:    true,
							Description: "When assigning entity types to targets, you can assign specific permissions to be synchronized using the `include_patterns`/`exclude_patterns` regular expressions.",
						},
					},
				},
				Required:    true,
				Description: "Target JPD",
			},
		},
		MarkdownDescription: "Provides a [JFrog Access Federation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/access-federation) resource to setup Star Topology.\n\n" +
			"~>The source and targets must have been configured properly for [Access Federation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/access-federation).\n\n" +
			"~>**Deletion** is currently not supported via REST API. This must be done using JFrog UI.",
	}
}

type accessFederationStarResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Entities types.Set    `tfsdk:"entities"`
	Targets  types.Set    `tfsdk:"targets"`
}

var targetAttributeTypes = map[string]attr.Type{
	"id":                 types.StringType,
	"url":                types.StringType,
	"permission_filters": types.ObjectType{AttrTypes: permissionFilterAttributeTypes},
}

var targetsElmementType = types.ObjectType{
	AttrTypes: targetAttributeTypes,
}

var permissionFilterAttributeTypes = map[string]attr.Type{
	"include_patterns": types.SetType{ElemType: types.StringType},
	"exclude_patterns": types.SetType{ElemType: types.StringType},
}

func (r *accessFederationStarResourceModel) fromAPIModel(ctx context.Context, apiModel *accessFederationGetResponseAPIModel) (ds diag.Diagnostics) {
	r.Targets = types.SetNull(targetsElmementType)

	if len(apiModel.Targets) > 0 {
		targets := lo.Map(
			apiModel.Targets,
			func(target accessFederationTargetAPIModel, _ int) attr.Value {
				includePatterns := types.SetNull(types.StringType)
				if len(target.PermissionFilters.IncludePatterns) > 0 {
					p, d := types.SetValueFrom(ctx, types.StringType, target.PermissionFilters.IncludePatterns)
					if d.HasError() {
						ds.Append(d...)
					}
					includePatterns = p
				}

				excludePatterns := types.SetNull(types.StringType)
				if len(target.PermissionFilters.ExcludePatterns) > 0 {
					p, d := types.SetValueFrom(ctx, types.StringType, target.PermissionFilters.ExcludePatterns)
					if d.HasError() {
						ds.Append(d...)
					}
					excludePatterns = p
				}

				permissionFilters, d := types.ObjectValue(
					permissionFilterAttributeTypes,
					map[string]attr.Value{
						"include_patterns": includePatterns,
						"exclude_patterns": excludePatterns,
					},
				)
				if d.HasError() {
					ds.Append(d...)
				}

				t, d := types.ObjectValue(
					targetAttributeTypes,
					map[string]attr.Value{
						"id":                 types.StringValue(target.ID),
						"url":                types.StringValue(target.URL),
						"permission_filters": permissionFilters,
					},
				)
				if d.HasError() {
					ds.Append(d...)
				}

				return t
			},
		)
		targetsSet, d := types.SetValue(targetsElmementType, targets)
		if d.HasError() {
			ds.Append(d...)
		}
		r.Targets = targetsSet
	}

	entities, d := types.SetValueFrom(ctx, types.StringType, apiModel.Entities)
	if d.HasError() {
		ds.Append(d...)
	}
	r.Entities = entities

	return
}

func (r accessFederationStarResourceModel) toAPIModel(ctx context.Context, apiModel *accessFederationRequestAPIModel) diag.Diagnostics {
	ds := diag.Diagnostics{}

	var entities []string
	ds.Append(r.Entities.ElementsAs(ctx, &entities, false)...)

	targets := lo.Map(
		r.Targets.Elements(),
		func(elem attr.Value, _ int) accessFederationTargetAPIModel {
			attrs := elem.(types.Object).Attributes()

			permissionFiltersAttrs := attrs["permission_filters"].(types.Object).Attributes()

			var includePatterns []string
			d := permissionFiltersAttrs["include_patterns"].(types.Set).ElementsAs(ctx, &includePatterns, false)
			if d.HasError() {
				ds.Append(d...)
			}

			var excludePatterns []string
			d = permissionFiltersAttrs["exclude_patterns"].(types.Set).ElementsAs(ctx, &excludePatterns, false)
			if d.HasError() {
				ds.Append(d...)
			}

			return accessFederationTargetAPIModel{
				ID:  attrs["id"].(types.String).ValueString(),
				URL: attrs["url"].(types.String).ValueString(),
				PermissionFilters: accessFederationPermissionFiltersAPIModel{
					IncludePatterns: includePatterns,
					ExcludePatterns: excludePatterns,
				},
			}
		},
	)

	*apiModel = accessFederationRequestAPIModel{
		ID:       r.ID.ValueString(),
		Entities: entities,
		Targets:  targets,
	}

	return ds
}

type accessFederationRequestAPIModel struct {
	ID       string                           `json:"id"`
	Entities []string                         `json:"entities"`
	Targets  []accessFederationTargetAPIModel `json:"targets"`
}

type accessFederationTargetAPIModel struct {
	ID                string                                    `json:"id"`
	URL               string                                    `json:"url"`
	PermissionFilters accessFederationPermissionFiltersAPIModel `json:"permission_filters"`
}

type accessFederationPermissionFiltersAPIModel struct {
	IncludePatterns []string `json:"include_patterns"`
	ExcludePatterns []string `json:"exclude_patterns"`
}

type accessFederationResponseAPIModel struct {
	Label  string `json:"label"`
	Status string `json:"status"`
}

type accessFederationGetResponseAPIModel struct {
	Entities []string                         `json:"entities"`
	Targets  []accessFederationTargetAPIModel `json:"targets"`
}

func (r *accessFederationStarResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *accessFederationStarResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan accessFederationStarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var accessFederation accessFederationRequestAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &accessFederation)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var results []accessFederationResponseAPIModel
	response, err := r.ProviderData.Client.R().
		SetPathParam("id", plan.ID.ValueString()).
		SetBody(accessFederation).
		SetResult(&results).
		Put(accessFederationEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	for _, result := range results {
		tflog.Info(ctx, "Create result", map[string]interface{}{
			"label":  result.Label,
			"status": result.Status,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *accessFederationStarResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state accessFederationStarResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var accessFederation accessFederationGetResponseAPIModel
	response, err := r.ProviderData.Client.R().
		SetPathParam("id", state.ID.ValueString()).
		SetResult(&accessFederation).
		Get(accessFederationEndpoint)

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.fromAPIModel(ctx, &accessFederation)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *accessFederationStarResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan accessFederationStarResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var accessFederation accessFederationRequestAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &accessFederation)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var results []accessFederationResponseAPIModel
	response, err := r.ProviderData.Client.R().
		SetPathParam("id", plan.ID.ValueString()).
		SetBody(accessFederation).
		SetResult(&results).
		Put(accessFederationEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	for _, result := range results {
		tflog.Info(ctx, "Update result", map[string]interface{}{
			"label":  result.Label,
			"status": result.Status,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *accessFederationStarResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	resp.Diagnostics.AddWarning(
		"Access Federation deletion not supported",
		" The resource has be deleted from Terraform state. To delete Access Federation relationship, please use the JFrog UI.",
	)

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *accessFederationStarResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
