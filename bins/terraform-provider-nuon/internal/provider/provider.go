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

// nuonProvider is the Nuon Terraform provider.
type nuonProvider struct {
	version string
}

// providerConfig is the resolved provider-level configuration handed to each
// resource via Configure.
type providerConfig struct {
	apiURL string
}

var _ provider.Provider = (*nuonProvider)(nil)

// New returns a provider factory for the given build version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &nuonProvider{version: version}
	}
}

func (p *nuonProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nuon"
	resp.Version = p.version
}

func (p *nuonProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provision Nuon install stacks from Terraform.",
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

func (p *nuonProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
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

func (p *nuonProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewStackResource,
	}
}

func (p *nuonProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
