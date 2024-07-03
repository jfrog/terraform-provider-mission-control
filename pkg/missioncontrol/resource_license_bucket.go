package missioncontrol

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validator_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const (
	licenseBucketsEndpoint = "mc/api/v1/buckets"
	licenseBucketEndpoint  = "mc/api/v1/buckets/{name}"
)

var _ resource.Resource = &licenseBucketResource{}

type licenseBucketResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

func NewLicenseBucketResource() resource.Resource {
	return &licenseBucketResource{
		TypeName: "missioncontrol_license_bucket",
	}
}

func (r *licenseBucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *licenseBucketResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier of this license bucket.",
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Name of the license bucket",
			},
			"url": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validator_string.IsURLHttpOrHttps(),
					stringvalidator.ConflictsWith(path.MatchRoot("file")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Signed URL of the license bucket. Can't be set together with `file`.",
			},
			"file": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ConflictsWith(path.MatchRoot("url")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "File path to the license bucket. Can't be set together with `url`.",
			},
			"key": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "License bucket key.",
			},
			"subject": schema.StringAttribute{
				Computed:    true,
				Description: "The customer name of this license bucket.",
			},
			"product_name": schema.StringAttribute{
				Computed: true,
			},
			"product_id": schema.Int64Attribute{
				Computed: true,
			},
			"license_type": schema.StringAttribute{
				Computed:    true,
				Description: "The license type of this license bucket.",
			},
			"issued_date": schema.StringAttribute{
				Computed:    true,
				Description: "The issue date for this license bucket.",
			},
			"valid_date": schema.StringAttribute{
				Computed:    true,
				Description: "The expiry date for this license bucket.",
			},
			"quantity": schema.Int64Attribute{
				Computed:    true,
				Description: "The total number of licenses in this bucket.",
			},
			"signature": schema.StringAttribute{
				Computed: true,
			},
			"used": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of used licenses in this bucket.",
			},
		},
		MarkdownDescription: "Provides a JFrog [license bucket](https://jfrog.com/help/r/jfrog-platform-administration-documentation/manage-license-buckets) resource to manage license buckets.",
	}
}

type licenseBucketResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	URL         types.String `tfsdk:"url"`
	File        types.String `tfsdk:"file"`
	Key         types.String `tfsdk:"key"`
	Subject     types.String `tfsdk:"subject"`
	ProductName types.String `tfsdk:"product_name"`
	ProductID   types.Int64  `tfsdk:"product_id"`
	LicenseType types.String `tfsdk:"license_type"`
	IssuedDate  types.String `tfsdk:"issued_date"`
	ValidDate   types.String `tfsdk:"valid_date"`
	Signature   types.String `tfsdk:"signature"`
	Quantity    types.Int64  `tfsdk:"quantity"`
	Used        types.Int64  `tfsdk:"used"`
}

func (r *licenseBucketResourceModel) fromAPIModel(_ context.Context, apiModel *licenseBucketPostResponseAPIModel) (ds diag.Diagnostics) {
	r.ID = types.StringValue(apiModel.ID)
	r.Name = types.StringValue(apiModel.Name)
	r.Subject = types.StringValue(apiModel.Subject)
	r.ProductName = types.StringValue(apiModel.ProductName)
	r.ProductID = types.Int64Value(apiModel.ProductID)
	r.LicenseType = types.StringValue(apiModel.LicenseType)
	r.IssuedDate = types.StringValue(apiModel.IssuedDate)
	r.ValidDate = types.StringValue(apiModel.ValidDate)
	r.Quantity = types.Int64Value(apiModel.Quantity)
	r.Signature = types.StringValue(apiModel.Signature)
	r.Used = types.Int64Value(apiModel.Used)

	return
}

type licenseBucketPostRequestAPIModel struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Key  string `json:"key"`
}

type licenseBucketPostResponseAPIModel struct {
	ID          string `json:"identifier"`
	Subject     string `json:"subject"`
	ProductName string `json:"product_name"`
	ProductID   int64  `json:"product_id"`
	LicenseType string `json:"license_type"`
	IssuedDate  string `json:"issued_date"`
	ValidDate   string `json:"valid_date"`
	Quantity    int64  `json:"quantity"`
	Signature   string `json:"signature"`
	Name        string `json:"name"`
	Used        int64  `json:"used"`
	URL         string `json:"url"`
}

type licenseBucketGetAPIModel struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Type       string `json:"license_type"`
}

func (r *licenseBucketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *licenseBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan licenseBucketResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result licenseBucketPostResponseAPIModel
	var response *resty.Response
	var err error
	if len(plan.URL.ValueString()) > 0 {
		license := licenseBucketPostRequestAPIModel{
			Name: plan.Name.ValueString(),
			URL:  plan.URL.ValueString(),
			Key:  plan.Key.ValueString(),
		}

		response, err = r.ProviderData.Client.R().
			SetBody(&license).
			SetResult(&result).
			Post(licenseBucketsEndpoint)
	} else {
		fileBytes, err := os.ReadFile(plan.File.ValueString())
		if err != nil {
			utilfw.UnableToCreateResourceError(resp, err.Error())
			return
		}

		response, err = r.ProviderData.Client.R().
			SetMultipartField(
				"file",
				filepath.Base(plan.File.ValueString()),
				"application/octet-stream",
				bytes.NewReader(fileBytes),
			).
			SetMultipartFormData(
				map[string]string{
					"name": plan.Name.ValueString(),
					"key":  plan.Key.ValueString(),
				},
			).
			SetResult(&result).
			Post(licenseBucketsEndpoint)

		if err != nil {
			utilfw.UnableToCreateResourceError(resp, err.Error())
			return
		}
	}

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

func (r *licenseBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state licenseBucketResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var licenseBuckets []licenseBucketGetAPIModel

	response, err := r.ProviderData.Client.R().
		SetResult(&licenseBuckets).
		Get(licenseBucketsEndpoint)

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	matchedBucket, ok := lo.Find(
		licenseBuckets,
		func(licenseBucket licenseBucketGetAPIModel) bool {
			return licenseBucket.Name == state.Name.ValueString()
		},
	)
	if !ok {
		utilfw.UnableToRefreshResourceError(resp, fmt.Sprintf("bucket %s can't be found", state.Name.ValueString()))
		return
	}

	state.Name = types.StringValue(matchedBucket.Name)
	state.Quantity = types.Int64Value(matchedBucket.Size)
	state.LicenseType = types.StringValue(matchedBucket.Type)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *licenseBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// noop
}

func (r *licenseBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state licenseBucketResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("name", state.Name.ValueString()).
		Delete(licenseBucketEndpoint)
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
