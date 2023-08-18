// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/powertoolsdev/mono/pkg/deprecated/api/gqlclient"
)

const (
	defaultAPIURL      string = "https://api.prod.nuon.co/graphql"
	apiTokenEnvVarName string = "NUON_API_TOKEN"
	apiURLEnvVarName   string = "NUON_API_URL"
	orgIDEnvVarName    string = "NUON_ORG_ID"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &Provider{}

// Provider defines the provider implementation.
type Provider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ProviderModel describes the provider data model.
type ProviderModel struct {
	APIURL       types.String `tfsdk:"api_url"`
	APIAuthToken types.String `tfsdk:"api_token"`
	OrgID        types.String `tfsdk:"org_id"`
}

type ProviderData struct {
	OrgID  string
	Client gqlclient.Client
}

func (p *Provider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nuon"
	resp.Version = p.version
}

func (p *Provider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				MarkdownDescription: "Override the url to use a custom endpoint.",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "API token to access the api.",
				Optional:            true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "Nuon org ID.",
				Optional:            true,
			},
		},
	}
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// set overrides using env vars
	apiURLEnvVar := os.Getenv(apiURLEnvVarName)
	if apiURLEnvVar != "" {
		data.APIURL = types.StringValue(apiURLEnvVar)
	}
	if data.APIURL.ValueString() == "" {
		data.APIURL = types.StringValue(defaultAPIURL)
	}

	orgIDEnvVar := os.Getenv(orgIDEnvVarName)
	if orgIDEnvVar != "" {
		data.OrgID = types.StringValue(orgIDEnvVar)
	}
	if data.OrgID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Org ID must be set",
			"Please set `org_id` on the provider, or the `NUON_ORG_ID` env var.",
		)
		return
	}

	apiTokenEnvVar := os.Getenv(apiTokenEnvVarName)
	if orgIDEnvVar != "" {
		data.APIAuthToken = types.StringValue(apiTokenEnvVar)
	}
	if data.APIAuthToken.ValueString() == "" {
		resp.Diagnostics.AddError(
			"api token must be set",
			"Please set `api_token` on the provider, or the `NUON_API_TOKEN` env var.",
		)
		return
	}

	v := validator.New()
	client, err := gqlclient.New(v,
		gqlclient.WithAuthToken(data.APIAuthToken.ValueString()),
		gqlclient.WithURL(data.APIURL.ValueString()),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"unable to initialize api client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	resp.DataSourceData = &ProviderData{
		Client: client,
		OrgID:  data.OrgID.ValueString(),
	}
	resp.ResourceData = &ProviderData{
		Client: client,
		OrgID:  data.OrgID.ValueString(),
	}
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAppResource,
		NewInstallResource,
		NewContainerImageComponentResource,
		NewDockerBuildComponentResource,
		NewHelmChartComponentResource,
		NewTerraformModuleComponentResource,
		NewDeployResource,
		NewBuildResource,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAppDataSource,
		NewOrgDataSource,
		NewConnectedRepoDataSource,
		NewConnectedReposDataSource,
		NewInstallDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}
