package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &ConnectedReposDataSource{}

func NewConnectedReposDataSource() datasource.DataSource {
	return &ConnectedReposDataSource{}
}

// ConnectedReposDataSource defines the data source implementation.
type ConnectedReposDataSource struct {
	baseDataSource
}

type ConnectedReposElementDataSource struct {
	// computed
	DefaultBranch types.String `tfsdk:"default_branch"`
	FullName      types.String `tfsdk:"full_name"`
	Repo          types.String `tfsdk:"repo"`
	Owner         types.String `tfsdk:"owner"`
	URL           types.String `tfsdk:"url"`
}

// ConnectedReposDataSourceModel describes the data source data model.
type ConnectedReposDataSourceModel struct {
	Repos []ConnectedReposElementDataSource `tfsdk:"repos"`
}

func (d *ConnectedReposDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connected_repos"
}

func (d *ConnectedReposDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Get all connected repos for your org.",
		MarkdownDescription: "Get all connected repos for your org.",

		Attributes: map[string]schema.Attribute{
			"repos": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
				},
			},
		},
	}
}

func (d *ConnectedReposDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConnectedReposDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "fetching repos for org id")
	repos, err := d.client.GetConnectedRepos(ctx, d.orgID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get repos",
			fmt.Sprintf("Please make sure your org_id (%s) is correct, and that the auth token has permissions for this org.", d.orgID),
		)
		return
	}

	for _, repo := range repos {
		data.Repos = append(data.Repos, ConnectedReposElementDataSource{
			DefaultBranch: types.StringValue(repo.DefaultBranch),
			FullName:      types.StringValue(repo.FullName),
			Repo:          types.StringValue(repo.Name),
			Owner:         types.StringValue(repo.Owner),
			URL:           types.StringValue(repo.Url),
		})
	}
	tflog.Trace(ctx, fmt.Sprintf("returned %d repos", len(repos)))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
