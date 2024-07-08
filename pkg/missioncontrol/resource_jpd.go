package missioncontrol

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validator_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const (
	jpdsEndpoint = "mc/api/v1/jpds"
	jpdEndpoint  = "mc/api/v1/jpds/{id}"
)

var _ resource.Resource = &jpdResource{}

type jpdResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

func NewJPDResource() resource.Resource {
	return &jpdResource{
		TypeName: "missioncontrol_jpd",
	}
}

func (r *jpdResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *jpdResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "A unique logical name for this Platform Deployment",
			},
			"url": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validator_string.IsURLHttpOrHttps(),
					stringvalidator.RegexMatches(regexp.MustCompile(`^.+/$`), "must end in '/'"),
				},
				Description: "The Platform deployment URL: http://<hostname>:<port>/; for example: http://myplatformserver:8082/. Note: For legacy instances, version 6.x and lower, the URL should contain the instance root context: http://<hostname>:<port>/<context>/; for example http://myv6server:8081/artifactory/. URL must ends with trailing slash.",
			},
			"token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ConflictsWith(path.MatchRoot("username"), path.MatchRoot("password")),
				},
				Description: "JPD join key",
			},
			"username": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(path.MatchRoot("password")),
					stringvalidator.ConflictsWith(path.MatchRoot("url")),
				},
				Description: "Admin username for legacy JPD (Artifactory 6.x).",
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(path.MatchRoot("username")),
					stringvalidator.ConflictsWith(path.MatchRoot("url")),
				},
				Description: "Admin password for legacy JPD (Artifactory 6.x).",
			},
			"location": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"city_name": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"country_code": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(2, 2),
						},
						Description: "2 letters ISO-3166-2 country code",
					},
					"latitude": schema.Float64Attribute{
						Required: true,
					},
					"longitude": schema.Float64Attribute{
						Required: true,
					},
				},
				Required:    true,
				Description: "The geographical location of the Platform Deployment to be displayed on a global Platform Deployment view",
			},
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Add labels to be applied for filtering Platform Deployments according to categories for example, location, dedicated centers - dev, testing, production",
			},
			"base_url": schema.StringAttribute{
				Computed: true,
			},
			"licenses": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"expired": schema.BoolAttribute{
							Computed: true,
						},
						"license_hash": schema.StringAttribute{
							Computed: true,
						},
						"licensed_to": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"valid_through": schema.StringAttribute{
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"services": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"code": schema.StringAttribute{
									Computed: true,
								},
							},
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"status": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"code": schema.StringAttribute{
						Computed: true,
					},
					"message": schema.StringAttribute{
						Computed: true,
					},
					"warnings": schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
				Computed: true,
			},
			"local": schema.BoolAttribute{
				Computed: true,
			},
			"is_cold_storage": schema.BoolAttribute{
				Computed: true,
			},
			"cold_storage_jpd": schema.StringAttribute{
				Computed: true,
			},
		},
		MarkdownDescription: "Provides a [JFrog Platform Deployment](https://jfrog.com/help/r/jfrog-platform-administration-documentation/manage-platform-deployments) resource to manage JPD.\n~>Supported on the Self-Hosted platform, with an Enterprise X or Enterprise+ license.",
	}
}

type jpdResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	URL            types.String `tfsdk:"url"`
	BaseURL        types.String `tfsdk:"base_url"`
	Token          types.String `tfsdk:"token"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	Location       types.Object `tfsdk:"location"`
	Services       types.Set    `tfsdk:"services"`
	Licenses       types.Set    `tfsdk:"licenses"`
	Tags           types.Set    `tfsdk:"tags"`
	Local          types.Bool   `tfsdk:"local"`
	Status         types.Object `tfsdk:"status"`
	IsColdStorage  types.Bool   `tfsdk:"is_cold_storage"`
	ColdStorageJPD types.String `tfsdk:"cold_storage_jpd"`
}

var licenseAttrTypes = map[string]attr.Type{
	"expired":       types.BoolType,
	"license_hash":  types.StringType,
	"licensed_to":   types.StringType,
	"type":          types.StringType,
	"valid_through": types.StringType,
}

var licenseElemType = types.ObjectType{
	AttrTypes: licenseAttrTypes,
}

var serviceAttrTypes = map[string]attr.Type{
	"status": types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"code": types.StringType,
		},
	},
	"type": types.StringType,
}

var serviceElemType = types.ObjectType{
	AttrTypes: serviceAttrTypes,
}

func (r *jpdResourceModel) fromAPIModel(ctx context.Context, apiModel *jpdGetResponseAPIModel) (ds diag.Diagnostics) {
	r.ID = types.StringValue(apiModel.ID)
	r.Name = types.StringValue(apiModel.Name)
	r.URL = types.StringValue(apiModel.URL)
	r.BaseURL = types.StringValue(apiModel.BaseURL)

	location, d := types.ObjectValue(
		map[string]attr.Type{
			"city_name":    types.StringType,
			"country_code": types.StringType,
			"latitude":     types.Float64Type,
			"longitude":    types.Float64Type,
		},
		map[string]attr.Value{
			"city_name":    types.StringValue(apiModel.Location.CityName),
			"country_code": types.StringValue(apiModel.Location.CountryCode),
			"latitude":     types.Float64Value(apiModel.Location.Latitude),
			"longitude":    types.Float64Value(apiModel.Location.Longitude),
		},
	)
	if d.HasError() {
		ds.Append(d...)
	}
	r.Location = location

	licenses := lo.Map(
		apiModel.Licenses,
		func(license jpdLicenseAPIModel, _ int) attr.Value {
			l, d := types.ObjectValue(
				licenseAttrTypes,
				map[string]attr.Value{
					"expired":       types.BoolValue(license.Expired),
					"license_hash":  types.StringValue(license.LicenseHash),
					"licensed_to":   types.StringValue(license.LicensedTo),
					"type":          types.StringValue(license.Type),
					"valid_through": types.StringValue(license.ValidThrough),
				},
			)
			if d.HasError() {
				ds.Append(d...)
			}

			return l
		},
	)
	licensesSet, d := types.SetValue(licenseElemType, licenses)
	if d.HasError() {
		ds.Append(d...)
	}
	r.Licenses = licensesSet

	services := lo.Map(
		apiModel.Services,
		func(service jpdServiceAPIModel, _ int) attr.Value {
			status, d := types.ObjectValue(
				map[string]attr.Type{
					"code": types.StringType,
				},
				map[string]attr.Value{
					"code": types.StringValue(service.Status.Code),
				},
			)
			if d.HasError() {
				ds.Append(d...)
			}

			s, d := types.ObjectValue(
				serviceAttrTypes,
				map[string]attr.Value{
					"status": status,
					"type":   types.StringValue(service.Type),
				},
			)
			if d.HasError() {
				ds.Append(d...)
			}

			return s
		},
	)
	servicesSet, d := types.SetValue(serviceElemType, services)
	if d.HasError() {
		ds.Append(d...)
	}
	r.Services = servicesSet

	warnings, d := types.SetValueFrom(ctx, types.StringType, apiModel.Status.Warnings)
	if d.HasError() {
		ds.Append(d...)
	}
	status, d := types.ObjectValue(
		map[string]attr.Type{
			"code":     types.StringType,
			"message":  types.StringType,
			"warnings": types.SetType{ElemType: types.StringType},
		},
		map[string]attr.Value{
			"code":     types.StringValue(apiModel.Status.Code),
			"message":  types.StringValue(apiModel.Status.Message),
			"warnings": warnings,
		},
	)
	if d.HasError() {
		ds.Append(d...)
	}
	r.Status = status

	tags, d := types.SetValueFrom(ctx, types.StringType, apiModel.Tags)
	if d.HasError() {
		ds.Append(d...)
	}
	r.Tags = tags
	r.Local = types.BoolValue(apiModel.Local)
	r.IsColdStorage = types.BoolValue(apiModel.IsColdStorage)

	r.ColdStorageJPD = types.StringNull()
	if apiModel.IsColdStorage {
		r.ColdStorageJPD = types.StringValue(apiModel.ColdStorageJPD)
	}

	return
}

func (r jpdResourceModel) toAPIModel(ctx context.Context, apiModel *jpdPostRequestAPIModel, artifactoryVersion string) diag.Diagnostics {
	ds := diag.Diagnostics{}

	var tags []string
	ds.Append(r.Tags.ElementsAs(ctx, &tags, false)...)

	locationAttrs := r.Location.Attributes()
	*apiModel = jpdPostRequestAPIModel{
		Name: r.Name.ValueString(),
		URL:  r.URL.ValueString(),
		Location: jpdLocationAPIModel{
			CityName:    locationAttrs["city_name"].(types.String).ValueString(),
			CountryCode: locationAttrs["country_code"].(types.String).ValueString(),
			Latitude:    locationAttrs["latitude"].(types.Float64).ValueFloat64(),
			Longitude:   locationAttrs["longitude"].(types.Float64).ValueFloat64(),
		},
		Tags: tags,
	}

	notLegacy, err := util.CheckVersion(artifactoryVersion, "7.0.0")
	if err != nil {
		ds.AddError("faild to check version", err.Error())
	}

	if notLegacy {
		apiModel.Token = r.Token.ValueString()
	} else {
		apiModel.Username = r.Username.ValueString()
		apiModel.Password = r.Password.ValueString()
	}

	return ds
}

type jpdPostRequestAPIModel struct {
	Name     string              `json:"name"`
	URL      string              `json:"url"`
	Token    string              `json:"token,omitempty"`
	Username string              `json:"username,omitempty"` // legacy (Artifactory 6.x only)
	Password string              `json:"password,omitempty"` // legacy (Artifactory 6.x only)
	Location jpdLocationAPIModel `json:"location"`
	Tags     []string            `json:"tags"`
}

type jpdLocationAPIModel struct {
	CityName    string  `json:"city_name"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type jpdGetResponseAPIModel struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	URL            string               `json:"url"`
	BaseURL        string               `json:"base_url"`
	Location       jpdLocationAPIModel  `json:"location"`
	Local          bool                 `json:"local"`
	Licenses       []jpdLicenseAPIModel `json:"licenses"`
	Services       []jpdServiceAPIModel `json:"services"`
	Status         jpdStatusAPIModel    `json:"status"`
	Tags           []string             `json:"tags"`
	IsColdStorage  bool                 `json:"is_cold_storage"`
	ColdStorageJPD string               `json:"cold_storage_jpd"`
}

type jpdLicenseAPIModel struct {
	Expired      bool   `json:"expired"`
	LicenseHash  string `json:"license_hash"`
	LicensedTo   string `json:"licensed_to"`
	Type         string `json:"type"`
	ValidThrough string `json:"valid_through"`
}

type jpdServiceAPIModel struct {
	Status jpdServiceStatusAPIModel `json:"status"`
	Type   string                   `json:"type"`
}

type jpdServiceStatusAPIModel struct {
	Code string `json:"code"`
}

type jpdStatusAPIModel struct {
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Warnings []string `json:"warnings"`
}

func (r *jpdResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *jpdResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan jpdResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jpd jpdPostRequestAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &jpd, r.ProviderData.ArtifactoryVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result jpdGetResponseAPIModel
	response, err := r.ProviderData.Client.R().
		SetBody(jpd).
		SetResult(&result).
		Post(jpdsEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(plan.fromAPIModel(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jpdResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state jpdResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jpd jpdGetResponseAPIModel
	response, err := r.ProviderData.Client.R().
		SetPathParam("id", state.ID.ValueString()).
		SetResult(&jpd).
		Get(jpdEndpoint)

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
	resp.Diagnostics.Append(state.fromAPIModel(ctx, &jpd)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *jpdResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan jpdResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state jpdResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jpd jpdPostRequestAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &jpd, r.ProviderData.ArtifactoryVersion)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("id", state.ID.ValueString()).
		SetBody(jpd).
		Put(jpdEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	var result jpdGetResponseAPIModel
	response, err = r.ProviderData.Client.R().
		SetPathParam("id", state.ID.ValueString()).
		SetResult(&result).
		Get(jpdEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(plan.fromAPIModel(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *jpdResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state jpdResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("id", state.ID.ValueString()).
		Delete(jpdEndpoint)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}
	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *jpdResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
