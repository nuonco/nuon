package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultAPIURL = "https://api.nuon.co"

// stackProvider is the Nuon Terraform provider.
type stackProvider struct {
	version string
}

// providerConfig is the resolved provider-level configuration handed to each
// resource via Configure.
type providerConfig struct {
	apiURL string
}

var _ provider.Provider = (*stackProvider)(nil)

// New returns a provider factory for the given build version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &stackProvider{version: version}
	}
}

func (p *stackProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "stack"
	resp.Version = p.version
}

func (p *stackProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Nuon install-stack configuration and report run status from Terraform.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Base URL of the Nuon API, up to but excluding `/v1`. Defaults to `" + defaultAPIURL + "`.",
			},
		},
	}
}

type providerModel struct {
	APIURL types.String `tfsdk:"api_url"`
}

func (p *stackProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiURL := defaultAPIURL
	if !data.APIURL.IsNull() && data.APIURL.ValueString() != "" {
		apiURL = data.APIURL.ValueString()
	}

	cfg := &providerConfig{apiURL: apiURL}
	resp.ResourceData = cfg
	resp.DataSourceData = cfg
}

func (p *stackProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPhoneHomeResource,
	}
}

func (p *stackProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewStackDataSource,
	}
}
