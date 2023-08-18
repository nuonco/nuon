package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &ConnectedRepoDataSource{}

func NewConnectedRepoDataSource() datasource.DataSource {
	return &ConnectedRepoDataSource{}
}

// ConnectedRepoDataSource defines the data source implementation.
type ConnectedRepoDataSource struct {
	baseDataSource
}

// ConnectedRepoDataSourceModel describes the data source data model.
type ConnectedRepoDataSourceModel struct {
	// inputs
	Name types.String `tfsdk:"name"`

	// computed
	DefaultBranch types.String `tfsdk:"default_branch"`
	FullName      types.String `tfsdk:"full_name"`
	Repo          types.String `tfsdk:"repo"`
	Owner         types.String `tfsdk:"owner"`
	URL           types.String `tfsdk:"url"`
}

func (d *ConnectedRepoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connected_repo"
}

func (d *ConnectedRepoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get a connected repo tied to your org.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name or URL of connected repo",
				Required:    true,
			},
			"default_branch": schema.StringAttribute{
				Computed: true,
			},
			"full_name": schema.StringAttribute{
				Computed: true,
			},
			"repo": schema.StringAttribute{
				Description: "this is the attribute to link to a connected config",
				Computed:    true,
			},
			"owner": schema.StringAttribute{
				Computed: true,
			},
			"url": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ConnectedRepoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConnectedRepoDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "fetching connected repo")
	repoResp, err := d.client.GetConnectedRepo(ctx, d.orgID, data.Name.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get connected repo")
		return
	}

	data.DefaultBranch = types.StringValue(repoResp.DefaultBranch)
	data.FullName = types.StringValue(repoResp.FullName)
	data.Repo = types.StringValue(repoResp.Name)
	data.Owner = types.StringValue(repoResp.Owner)
	data.URL = types.StringValue(repoResp.Url)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
