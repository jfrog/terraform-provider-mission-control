package missioncontrol

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	"github.com/samber/lo"
)

const (
	accessFederationsEndpoint    = "mc/api/v1/federation"
	accessFederationMeshEndpoint = "mc/api/v1/federation/create_mesh"
)

var _ resource.Resource = &accessFederationMeshResource{}

type accessFederationMeshResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

func NewAccessFederationMeshResource() resource.Resource {
	return &accessFederationMeshResource{
		TypeName: "missioncontrol_access_federation_mesh",
	}
}

func (r *accessFederationMeshResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *accessFederationMeshResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{ // this is necessary because TF can't use SetAttribute for state comparison
				Computed: true,
			},
			"ids": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(2),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
				Description: "IDs for the source Platform Deployment. Use [Get Access Federation Candidate API](https://jfrog.com/help/r/jfrog-rest-apis/get-access-federation-candidates) to get a list of ID. Must have at least 2 items.",
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
		},
		MarkdownDescription: "Provides a [JFrog Access Federation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/access-federation) resource to setup Mesh Topology.\n" +
			"~>The source and targets must have been configured properly for [Access Federation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/access-federation).\n" +
			"~>**Deletion** is currently not supported via REST API. This must be done using JFrog UI.",
	}
}

type accessFederationMeshResourceModel struct {
	ID       types.String `tfsdk:"id"`
	IDs      types.Set    `tfsdk:"ids"`
	Entities types.Set    `tfsdk:"entities"`
}

func (r *accessFederationMeshResourceModel) fromAPIModel(ctx context.Context, apiModel *accessFederationGetAllResponseAPIModel) (ds diag.Diagnostics) {
	var ids []string

	ids = append(ids, apiModel.Source)

	targetIDs := lo.Map(
		apiModel.Targets,
		func(target accessFederationTargetGetAllAPIModel, _ int) string {
			return target.ID
		},
	)
	ids = append(ids, targetIDs...)

	idsSet, d := types.SetValueFrom(ctx, types.StringType, ids)
	if d.HasError() {
		ds.Append(d...)
	}
	r.IDs = idsSet

	r.ID = types.StringValue(strings.Join(ids, ":"))

	entitiesNested := lo.Map(
		apiModel.Targets,
		func(target accessFederationTargetGetAllAPIModel, _ int) []string {
			return target.Entities
		},
	)
	entities := lo.Uniq(lo.Flatten(entitiesNested))

	entitiesSet, d := types.SetValueFrom(ctx, types.StringType, entities)
	if d.HasError() {
		ds.Append(d...)
	}
	r.Entities = entitiesSet

	return
}

func (r accessFederationMeshResourceModel) toAPIModel(ctx context.Context, apiModel *accessFederationMeshRequestAPIModel) diag.Diagnostics {
	ds := diag.Diagnostics{}

	var ids []string
	ds.Append(r.IDs.ElementsAs(ctx, &ids, false)...)

	var entities []string
	ds.Append(r.Entities.ElementsAs(ctx, &entities, false)...)

	*apiModel = accessFederationMeshRequestAPIModel{
		IDs:      ids,
		Entities: entities,
	}

	return ds
}

type accessFederationMeshRequestAPIModel struct {
	IDs      []string `json:"jpd_ids"`
	Entities []string `json:"entities"`
}

type accessFederationGetAllResponseAPIModel struct {
	Source  string                                 `json:"source"`
	Targets []accessFederationTargetGetAllAPIModel `json:"targets"`
}

type accessFederationTargetGetAllAPIModel struct {
	accessFederationTargetAPIModel
	Entities []string `json:"entities"`
}

func (r *accessFederationMeshResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *accessFederationMeshResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan accessFederationMeshResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var accessFederation accessFederationMeshRequestAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &accessFederation)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var results []accessFederationResponseAPIModel
	response, err := r.ProviderData.Client.R().
		SetBody(accessFederation).
		SetResult(&results).
		Post(accessFederationMeshEndpoint)

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

	var ids []string
	resp.Diagnostics.Append(plan.IDs.ElementsAs(ctx, &ids, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(strings.Join(ids, ":"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *accessFederationMeshResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state accessFederationMeshResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var accessFederations []accessFederationGetAllResponseAPIModel
	response, err := r.ProviderData.Client.R().
		SetQueryParam("includeNonConfiguredJPDs", "false").
		SetResult(&accessFederations).
		Get(accessFederationsEndpoint)

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	var jpdIDs []string
	resp.Diagnostics.Append(state.IDs.ElementsAs(ctx, &jpdIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceAccessFederation, found := lo.Find(
		accessFederations,
		func(accessFederation accessFederationGetAllResponseAPIModel) bool {
			targetIDs := lo.Map(
				accessFederation.Targets,
				func(target accessFederationTargetGetAllAPIModel, _ int) string {
					return target.ID
				},
			)

			return lo.Contains(jpdIDs, accessFederation.Source) && lo.Every(jpdIDs, targetIDs)
		},
	)

	if !found {
		utilfw.UnableToRefreshResourceError(
			resp,
			fmt.Sprintf("unabled to find Access Federation Configurations for JPDs: %s", strings.Join(jpdIDs, ", ")),
		)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.fromAPIModel(ctx, &sourceAccessFederation)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *accessFederationMeshResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan accessFederationMeshResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var accessFederation accessFederationMeshRequestAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &accessFederation)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var results []accessFederationResponseAPIModel
	response, err := r.ProviderData.Client.R().
		SetBody(accessFederation).
		SetResult(&results).
		Post(accessFederationMeshEndpoint)

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

	var ids []string
	resp.Diagnostics.Append(plan.IDs.ElementsAs(ctx, &ids, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(strings.Join(ids, ":"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *accessFederationMeshResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	resp.Diagnostics.AddWarning(
		"Access Federation deletion not supported",
		" The resource has be deleted from Terraform state. To delete Access Federation relationship, please use the JFrog UI.",
	)

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *accessFederationMeshResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) <= 1 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			"Expected at least one JPD ID in the form of: jpd_id_1:jpd_id_2:...",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ids"), parts)...)
}
