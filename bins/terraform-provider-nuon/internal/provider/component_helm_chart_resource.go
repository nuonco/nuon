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
	"github.com/powertoolsdev/mono/pkg/api/client/models"
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

func (r *HelmChartComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *HelmChartComponentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compResp, err := r.restClient.CreateComponent(ctx, data.AppID.ValueString(), &models.ServiceCreateComponentRequest{
		Name: data.Name.ValueStringPointer(),
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component")
		return
	}
	tflog.Trace(ctx, "got ID -- "+compResp.ID)
	data.ID = types.StringValue(compResp.ID)

	configRequest := &models.ServiceCreateHelmComponentConfigRequest{
		ChartName:                data.ChartName.ValueStringPointer(),
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Values:                   map[string]string{},
	}
	if data.PublicRepo != nil {
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			// TODO
			// Branch:    data.PublicRepo.GitRef.ValueStringPointer(),
			Directory: data.PublicRepo.Directory.ValueStringPointer(),
			Repo:      data.PublicRepo.Repo.ValueStringPointer(),
		}
	} else {
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    data.ConnectedRepo.Branch.ValueString(),
			Directory: data.ConnectedRepo.Directory.ValueStringPointer(),
			GitRef:    data.ConnectedRepo.GitRef.ValueString(),
			Repo:      data.ConnectedRepo.Repo.ValueStringPointer(),
		}
	}
	for _, value := range data.Value {
		configRequest.Values[value.Name.String()] = value.Value.String()
	}
	_, err = r.restClient.CreateHelmComponentConfig(ctx, compResp.ID, configRequest)
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component config")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "successfully created component")
}

func (r *HelmChartComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *HelmChartComponentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compResp, err := r.restClient.GetComponent(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get component")
		return
	}
	data.Name = types.StringValue(compResp.Name)

	configResp, err := r.restClient.GetComponentLatestConfig(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get component config")
		return
	}
	data.ChartName = types.StringValue(configResp.Helm.ChartName)
	if configResp.Helm.PublicGitVcsConfig != nil {
		// TODO
		// data.PublicRepo.GitRef = types.StringValue(configResp.Helm.PublicGitVcsConfig.Branch)
		data.PublicRepo.Directory = types.StringValue(configResp.Helm.PublicGitVcsConfig.Directory)
		data.PublicRepo.Repo = types.StringValue(configResp.Helm.PublicGitVcsConfig.Repo)
	} else {
		data.ConnectedRepo.Branch = types.StringValue(configResp.Helm.ConnectedGithubVcsConfig.Branch)
		data.ConnectedRepo.Directory = types.StringValue(configResp.Helm.ConnectedGithubVcsConfig.Directory)
		// TODO
		// data.ConnectedRepo.GitRef = types.StringValue(configResp.Helm.ConnectedGithubVcsConfig.Branch)
		data.ConnectedRepo.Repo = types.StringValue(configResp.Helm.ConnectedGithubVcsConfig.Repo)
	}
	for key, val := range configResp.DockerBuild.EnvVars {
		data.Value = append(data.Value, HelmValue{
			Name:  types.StringValue(key),
			Value: types.StringValue(val),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "successfully read component")
}

func (r *HelmChartComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *HelmChartComponentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.restClient.DeleteComponent(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "delete component")
		return
	}

	if !deleted {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "delete component")
		return
	}

	tflog.Trace(ctx, "successfully deleted component")
}

func (r *HelmChartComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *HelmChartComponentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updating component "+data.ID.ValueString())

	compResp, err := r.restClient.UpdateComponent(ctx, data.ID.ValueString(), &models.ServiceUpdateComponentRequest{
		Name: data.Name.ValueStringPointer(),
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "update component")
		return
	}
	data.Name = types.StringValue(compResp.Name)

	configRequest := &models.ServiceCreateHelmComponentConfigRequest{
		ChartName:                data.ChartName.ValueStringPointer(),
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Values:                   map[string]string{},
	}
	if data.PublicRepo != nil {
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			// TODO
			// Branch:    data.PublicRepo.GitRef.ValueStringPointer(),
			Directory: data.PublicRepo.Directory.ValueStringPointer(),
			Repo:      data.PublicRepo.Repo.ValueStringPointer(),
		}
	} else {
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    data.ConnectedRepo.Branch.ValueString(),
			Directory: data.ConnectedRepo.Directory.ValueStringPointer(),
			GitRef:    data.ConnectedRepo.GitRef.ValueString(),
			Repo:      data.ConnectedRepo.Repo.ValueStringPointer(),
		}
	}
	for _, value := range data.Value {
		configRequest.Values[value.Name.String()] = value.Value.String()
	}
	_, err = r.restClient.CreateHelmComponentConfig(ctx, compResp.ID, configRequest)
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component config")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "successfully updated component")
}

func (r *HelmChartComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
