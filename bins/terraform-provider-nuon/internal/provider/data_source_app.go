package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/powertoolsdev/mono/pkg/api/gqlclient"
)

var _ datasource.DataSource = &AppDataSource{}

func NewAppDataSource() datasource.DataSource {
	return &AppDataSource{}
}

// AppDataSource defines the data source implementation.
type AppDataSource struct {
	client gqlclient.Client
}

// AppDataSourceModel describes the data source data model.
type AppDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

func (d *AppDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (d *AppDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "`nuon_app` provides information about a Nuon app.\nThis data source can be useful when adding components and installs to an app created in the UI.",
		MarkdownDescription: "`nuon_app` provides information about a Nuon app.\nThis data source can be useful when adding components and installs to an app created in the UI.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "app name",
				MarkdownDescription: "app name",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Description:         "App id.",
				MarkdownDescription: "App id.",
				Required:            true,
			},
		},
	}
}

func (d *AppDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(gqlclient.Client)
	if !ok {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, fmt.Errorf("error setting client"), "configure resource")
		return
	}

	d.client = client
}

func (d *AppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AppDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "fetching app by id")
	appResp, err := d.client.GetApp(ctx, data.Id.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get app")
		return
	}
	data.Name = types.StringValue(appResp.Name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
