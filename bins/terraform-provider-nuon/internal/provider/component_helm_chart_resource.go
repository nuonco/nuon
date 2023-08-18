package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/powertoolsdev/mono/pkg/deprecated/api/gqlclient"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &HelmChartComponentResource{}
var _ resource.ResourceWithImportState = &HelmChartComponentResource{}

func NewHelmChartComponentResource() resource.Resource {
	return &HelmChartComponentResource{}
}

// HelmChartComponentResource defines the resource implementation.
type HelmChartComponentResource struct {
	baseResource
}

// HelmChartComponentResourceModel describes the resource data model.
type HelmChartComponentResourceModel struct {
	ID types.String `tfsdk:"id"`

	Name      types.String `tfsdk:"name"`
	AppID     types.String `tfsdk:"app_id"`
	ChartName types.String `tfsdk:"chart_name"`

	ConnectedRepo *ConnectedRepo `tfsdk:"connected_repo"`
	PublicRepo    *PublicRepo    `tfsdk:"public_repo"`

	Value []HelmValue `tfsdk:"value"`
}

func (r *HelmChartComponentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_helm_chart_component"
}

func (r *HelmChartComponentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deploy any helm chart",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Component id",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Component name",
				Optional:            false,
				Required:            true,
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "ID of the app this component belongs too.",
				Optional:            false,
				Required:            true,
			},
			"chart_name": schema.StringAttribute{
				MarkdownDescription: "Name that the chart will get installed as.",
				Optional:            false,
				Required:            true,
			},
			"public_repo":    publicRepoAttribute(),
			"connected_repo": connectedRepoAttribute(),
		},
		Blocks: map[string]schema.Block{
			"value": helmValueSharedBlock(),
		},
	}
}

func (r *HelmChartComponentResource) getConfigInput(data *HelmChartComponentResourceModel) (*gqlclient.ComponentConfigInput, error) {
	vals := make([]*gqlclient.KeyValuePairInput, 0)
	for _, val := range data.Value {
		vals = append(vals, val.toKeyValueInput())
	}

	cfg := &gqlclient.ComponentConfigInput{
		BuildConfig: &gqlclient.BuildConfigInput{
			HelmBuildConfig: &gqlclient.HelmBuildInput{
				ChartName: data.ChartName.ValueString(),
			},
		},
		DeployConfig: &gqlclient.DeployConfigInput{
			HelmDeployConfig: &gqlclient.HelmDeployInput{
				Values: vals,
			},
		},
	}

	if data.PublicRepo != nil {
		cfg.BuildConfig.HelmBuildConfig.VcsConfig = data.PublicRepo.getVCSConfig()
	} else {
		cfg.BuildConfig.HelmBuildConfig.VcsConfig = data.ConnectedRepo.getVCSConfig()
	}

	return cfg, nil
}

func (r *HelmChartComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *HelmChartComponentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "creating component")
	cfgInput, err := r.getConfigInput(data)
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component")
		return
	}

	compResp, err := r.client.UpsertComponent(ctx, gqlclient.ComponentInput{
		Id:     data.ID.ValueString(),
		AppId:  data.AppID.ValueString(),
		Name:   data.Name.ValueString(),
		Config: cfgInput,
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component")
		return
	}
	data.ID = types.StringValue(compResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HelmChartComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *HelmChartComponentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compResp, err := r.client.GetComponent(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get component")
		return
	}
	data.Name = types.StringValue(compResp.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HelmChartComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *HelmChartComponentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.client.DeleteComponent(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "delete component")
		return
	}

	if !deleted {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "delete component")
		return
	}
}

func (r *HelmChartComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *HelmChartComponentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updating component "+data.ID.ValueString())
	cfgInput, err := r.getConfigInput(data)
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "update component")
		return
	}

	installResp, err := r.client.UpsertComponent(ctx, gqlclient.ComponentInput{
		AppId:  data.AppID.ValueString(),
		Id:     data.ID.ValueString(),
		Name:   data.Name.ValueString(),
		Config: cfgInput,
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "update component")
		return
	}

	data.Name = types.StringValue(installResp.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HelmChartComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
