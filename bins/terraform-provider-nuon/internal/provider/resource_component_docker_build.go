package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/powertoolsdev/mono/pkg/api/gqlclient"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DockerBuildComponentResource{}
var _ resource.ResourceWithImportState = &DockerBuildComponentResource{}

func NewDockerBuildComponentResource() resource.Resource {
	return &DockerBuildComponentResource{}
}

// DockerBuildComponentResource defines the resource implementation.
type DockerBuildComponentResource struct {
	client gqlclient.Client
}

// DockerBuildComponentResourceModel describes the resource data model.
type DockerBuildComponentResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	AppID types.String `tfsdk:"app_id"`

	SyncOnly    types.Bool   `tfsdk:"sync_only"`
	BasicDeploy *BasicDeploy `tfsdk:"basic_deploy"`
	EnvVar      []EnvVar     `tfsdk:"env_var"`

	Dockerfile    types.String   `tfsdk:"dockerfile"`
	ConnectedRepo *ConnectedRepo `tfsdk:"connected_repo"`
	PublicRepo    *PublicRepo    `tfsdk:"public_repo"`
}

func (r *DockerBuildComponentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_build_component"
}

func (r *DockerBuildComponentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Build any image in a connected or public github repo.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID",
				Computed:            true,
				Required:            false,
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
			"sync_only": schema.BoolAttribute{
				MarkdownDescription: "Set to true to only use this image for syncing (ie: no deployment).",
				Optional:            true,
				Required:            false,
			},
			"dockerfile": schema.StringAttribute{
				MarkdownDescription: "dockerfile",
				Optional:            true,
				Default:             stringdefault.StaticString("Dockerfile"),
				Computed:            true,
			},
			"public_repo":    publicRepoAttribute(),
			"connected_repo": connectedRepoAttribute(),
			"basic_deploy":   basicDeployAttribute(),
		},
		Blocks: map[string]schema.Block{
			"env_var": envVarSharedBlock(),
		},
	}
}

func (r *DockerBuildComponentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(gqlclient.Client)
	if !ok {
		writeDiagnosticsErr(ctx, resp.Diagnostics, fmt.Errorf("error setting client"), "configure resource")
		return
	}

	r.client = client
}
func (r *DockerBuildComponentResource) getConfigInput(data *DockerBuildComponentResourceModel) (*gqlclient.ComponentConfigInput, error) {
	envVars := make([]*gqlclient.KeyValuePairInput, 0)
	for _, envVar := range data.EnvVar {
		envVars = append(envVars, envVar.toKeyValueInput())
	}

	cfg := &gqlclient.ComponentConfigInput{
		BuildConfig: &gqlclient.BuildConfigInput{
			DockerBuildConfig: &gqlclient.DockerBuildInput{
				BuildArgs:     []*gqlclient.KeyValuePairInput{},
				Dockerfile:    data.Dockerfile.ValueString(),
				EnvVarsConfig: []*gqlclient.KeyValuePairInput{},
			},
		},
		DeployConfig: &gqlclient.DeployConfigInput{},
	}

	// handle deploy config
	if data.SyncOnly.ValueBool() {
		cfg.DeployConfig.Noop = true
	} else {
		cfg.DeployConfig = data.BasicDeploy.toDeployConfigInput()
		cfg.DeployConfig.BasicDeployConfig.EnvVars = envVars
	}

	if data.PublicRepo != nil {
		cfg.BuildConfig.DockerBuildConfig.VcsConfig = data.PublicRepo.getVCSConfig()
	} else {
		cfg.BuildConfig.DockerBuildConfig.VcsConfig = data.ConnectedRepo.getVCSConfig()
	}

	cfg.BuildConfig.DockerBuildConfig.VcsConfig = nil
	return cfg, nil
}

func (r *DockerBuildComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DockerBuildComponentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfgInput, err := r.getConfigInput(data)
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "get component config")
		return
	}

	compResp, err := r.client.UpsertComponent(ctx, gqlclient.ComponentInput{
		Id:     data.ID.ValueString(),
		AppId:  data.AppID.ValueString(),
		Name:   data.Name.ValueString(),
		Config: cfgInput,
	})
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "upsert component")
		return
	}
	data.ID = types.StringValue(compResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DockerBuildComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DockerBuildComponentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	compResp, err := r.client.GetComponent(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "get component")
		return
	}
	data.Name = types.StringValue(compResp.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DockerBuildComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DockerBuildComponentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.client.DeleteComponent(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "delete component")
		return
	}

	if !deleted {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "delete component")
		return
	}
}

func (r *DockerBuildComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DockerBuildComponentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updating component "+data.ID.ValueString())
	cfgInput, err := r.getConfigInput(data)
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "update component")
		return
	}

	installResp, err := r.client.UpsertComponent(ctx, gqlclient.ComponentInput{
		AppId:  data.AppID.ValueString(),
		Id:     data.ID.ValueString(),
		Name:   data.Name.ValueString(),
		Config: cfgInput,
	})
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "get app")
		return
	}

	data.Name = types.StringValue(installResp.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DockerBuildComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
