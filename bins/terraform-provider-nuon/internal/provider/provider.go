// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/powertoolsdev/mono/pkg/api/gqlclient"
)

const (
	defaultAPIURL string = "https://api.stage.nuon.co/graphql"
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
	APIAuthToken types.String `tfsdk:"api_auth_token"`
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
			"api_auth_token": schema.StringAttribute{
				MarkdownDescription: "Auth token to access the api.",
				Required:            true,
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

	apiURL := defaultAPIURL
	if !data.APIURL.IsNull() {
		apiURL = data.APIURL.ValueString()
	}

	v := validator.New()
	client, err := gqlclient.New(v,
		gqlclient.WithAuthToken(data.APIAuthToken.ValueString()),
		gqlclient.WithURL(apiURL),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"unable to initialize api client",
			"Please report this issue to the provider developers.",
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
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
