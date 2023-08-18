package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &OrgDataSource{}

func NewOrgDataSource() datasource.DataSource {
	return &OrgDataSource{}
}

// OrgDataSource defines the data source implementation.
type OrgDataSource struct {
	baseDataSource
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
				Computed:            true,
			},
		},
	}
}

func (d *OrgDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrgDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgResp, err := d.client.GetOrg(ctx, d.orgID)
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get org")
		return
	}
	data.Name = types.StringValue(orgResp.Name)
	data.Id = types.StringValue(d.orgID)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
