package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/powertoolsdev/mono/pkg/api/gqlclient"
)

var _ datasource.DataSource = &OrgDataSource{}

func NewOrgDataSource() datasource.DataSource {
	return &OrgDataSource{}
}

// OrgDataSource defines the data source implementation.
type OrgDataSource struct {
	client gqlclient.Client
}

// OrgDataSourceModel describes the data source data model.
type OrgDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

func (d *OrgDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_org"
}

func (d *OrgDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "`nuon_org` provides information about a Nuon org.\nThis data source can be useful when adding components and installs to an org created in the UI.",
		MarkdownDescription: "`nuon_org` provides information about a Nuon org.\nThis data source can be useful when adding components and installs to an org created in the UI.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "org name",
				MarkdownDescription: "org name",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Description:         "Org id.",
				MarkdownDescription: "Org id.",
				Required:            true,
			},
		},
	}
}

func (d *OrgDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(gqlclient.Client)
	if !ok {
		writeDiagnosticsErr(ctx, resp.Diagnostics, fmt.Errorf("error setting client"), "configure resource")
		return
	}

	d.client = client
}

func (d *OrgDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgResp, err := d.client.GetOrg(ctx, data.Id.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "get org")
		return
	}
	data.Name = types.StringValue(orgResp.Name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
