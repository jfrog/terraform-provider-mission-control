package missioncontrol

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
	validator_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
)

var Version = "1.0.0"

// needs to be exported so make file can update this
var productId = "terraform-provider-mission-control/" + Version

var _ provider.Provider = (*MissionControlProvider)(nil)

type MissionControlProvider struct {
	Meta util.ProviderMetadata
}

type missionControlProviderModel struct {
	URL              types.String `tfsdk:"url"`
	AccessToken      types.String `tfsdk:"access_token"`
	OIDCProviderName types.String `tfsdk:"oidc_provider_name"`
}

func NewProvider() func() provider.Provider {
	return func() provider.Provider {
		return &MissionControlProvider{}
	}
}

func (p *MissionControlProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Check environment variables, first available OS variable will be assigned to the var
	url := util.CheckEnvVars([]string{"JFROG_URL"}, "")
	accessToken := util.CheckEnvVars([]string{"JFROG_ACCESS_TOKEN"}, "")

	var config missionControlProviderModel

	// Read configuration data into model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.URL.ValueString() != "" {
		url = config.URL.ValueString()
	}

	if url == "" {
		resp.Diagnostics.AddError(
			"Missing URL Configuration",
			"While configuring the provider, the url was not found in the JFROG_URL environment variable or provider configuration block url attribute.",
		)
		return
	}

	platformClient, err := client.Build(url, productId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resty client",
			err.Error(),
		)
		return
	}

	oidcAccessToken, err := util.OIDCTokenExchange(ctx, platformClient, config.OIDCProviderName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed OIDC ID token exchange",
			err.Error(),
		)
		return
	}

	// use token from OIDC provider, which should take precedence over
	// environment variable data, if found.
	if oidcAccessToken != "" {
		accessToken = oidcAccessToken
	}

	// use token from configuration, which should take precedence over
	// environment variable data or OIDC provider, if found.
	if config.AccessToken.ValueString() != "" {
		accessToken = config.AccessToken.ValueString()
	}

	if accessToken == "" {
		resp.Diagnostics.AddWarning(
			"Missing JFrog Access Token",
			"Access Token was not found in the JFROG_ACCESS_TOKEN environment variable, provider configuration block access_token attribute, or Terraform Cloud TFC_WORKLOAD_IDENTITY_TOKEN environment variable. Mission Control functionality will be affected.",
		)
	}

	_, err = client.AddAuth(platformClient, "", accessToken)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Auth to Resty client",
			err.Error(),
		)
		return
	}

	version, err := util.GetArtifactoryVersion(platformClient)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Artifactory version",
			err.Error(),
		)
		return
	}

	featureUsage := fmt.Sprintf("Terraform/%s", req.TerraformVersion)
	go util.SendUsage(ctx, platformClient.R(), productId, featureUsage)

	meta := util.ProviderMetadata{
		Client:             platformClient,
		ArtifactoryVersion: version,
	}

	p.Meta = meta

	resp.DataSourceData = meta
	resp.ResourceData = meta
}

func (p *MissionControlProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "missioncontrol"
	resp.Version = Version
}

func (p *MissionControlProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewDataSource,
	}
}

func (p *MissionControlProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewLicenseBucketResource,
		NewJPDResource,
	}
}

func (p *MissionControlProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					validator_string.IsURLHttpOrHttps(),
				},
				MarkdownDescription: "JFrog Platform URL. This can also be sourced from the `JFROG_URL` environment variable.",
			},
			"access_token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "This is a access token that can be given to you by your admin under `Platform Configuration -> User Management -> Access Tokens`. This can also be sourced from the `JFROG_ACCESS_TOKEN` environment variable.",
			},
			"oidc_provider_name": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "OIDC provider name. See [Configure an OIDC Integration](https://jfrog.com/help/r/jfrog-platform-administration-documentation/configure-an-oidc-integration) for more details.",
			},
		},
		MarkdownDescription: "The JFrog Mission Control provider provides resources to interact with Mission Control supported by JFrog Platform. See [official documentation](https://jfrog.com/help/r/get-started-with-the-jfrog-platform/jfrog-mission-control) for more details.",
	}
}
