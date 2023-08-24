package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DockerBuildComponentResource{}
var _ resource.ResourceWithImportState = &DockerBuildComponentResource{}

func NewDockerBuildComponentResource() resource.Resource {
	return &DockerBuildComponentResource{}
}

// DockerBuildComponentResource defines the resource implementation.
type DockerBuildComponentResource struct {
	baseResource
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

func (r *DockerBuildComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DockerBuildComponentResourceModel
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

	configRequest := &models.ServiceCreateDockerBuildComponentConfigRequest{
		BuildArgs:  []string{},
		Dockerfile: data.Dockerfile.ValueString(),
		SyncOnly:   data.SyncOnly.ValueBool(),
		Target:     "",
		EnvVars:    map[string]string{},
	}
	if data.BasicDeploy != nil {
		configRequest.BasicDeployConfig = &models.ServiceBasicDeployConfigRequest{
			Args:            []string{},
			CPULimit:        "",
			CPURequest:      "",
			EnvVars:         map[string]string{},
			HealthCheckPath: data.BasicDeploy.HealthCheckPath.String(),
			InstanceCount:   data.BasicDeploy.InstanceCount.ValueInt64(),
			ListenPort:      data.BasicDeploy.Port.ValueInt64(),
			MemLimit:        "",
			MemRequest:      "",
		}
	}
	if data.PublicRepo != nil {
		public := data.PublicRepo
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    public.Branch.ValueStringPointer(),
			Directory: public.Directory.ValueStringPointer(),
			Repo:      public.Repo.ValueStringPointer(),
		}
	} else {
		connected := data.ConnectedRepo
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    connected.Branch.ValueString(),
			Directory: connected.Directory.ValueStringPointer(),
			Repo:      connected.Repo.ValueStringPointer(),
		}
	}
	for _, envVar := range data.EnvVar {
		configRequest.EnvVars[envVar.Name.String()] = envVar.Value.String()
		if configRequest.BasicDeployConfig != nil {
			configRequest.BasicDeployConfig.EnvVars[envVar.Name.String()] = envVar.Value.String()
		}
	}
	_, err = r.restClient.CreateDockerBuildComponentConfig(ctx, data.ID.ValueString(), configRequest)
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component config")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "successfully created component")
}

func (r *DockerBuildComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DockerBuildComponentResourceModel

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
	data.AppID = types.StringValue(compResp.AppID)

	configResp, err := r.restClient.GetComponentLatestConfig(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get component config")
		return
	}
	if configResp.DockerBuild == nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, errors.New("did not get docker build config"), "get component config")
		return
	}
	dockerBuild := configResp.DockerBuild
	data.Dockerfile = types.StringValue(dockerBuild.Dockerfile)
	data.SyncOnly = types.BoolValue(dockerBuild.SyncOnly)
	if dockerBuild.BasicDeployConfig != nil {
		basicDeployConfig := dockerBuild.BasicDeployConfig
		data.BasicDeploy = &BasicDeploy{
			HealthCheckPath: types.StringValue(basicDeployConfig.HealthCheckPath),
			InstanceCount:   types.Int64Value(basicDeployConfig.InstanceCount),
			Port:            types.Int64Value(basicDeployConfig.ListenPort),
		}
	}
	if dockerBuild.ConnectedGithubVcsConfig != nil {
		connected := dockerBuild.ConnectedGithubVcsConfig
		data.ConnectedRepo = &ConnectedRepo{
			Branch:    types.StringValue(connected.Branch),
			Directory: types.StringValue(connected.Directory),
			Repo:      types.StringValue(connected.Repo),
		}
	} else {
		public := dockerBuild.PublicGitVcsConfig
		data.PublicRepo = &PublicRepo{
			Branch:    types.StringValue(public.Branch),
			Directory: types.StringValue(public.Directory),
			Repo:      types.StringValue(public.Repo),
		}
	}
	for key, val := range configResp.DockerBuild.EnvVars {
		data.EnvVar = append(data.EnvVar, EnvVar{
			Name:  types.StringValue(key),
			Value: types.StringValue(val),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "successfully read component")
}

func (r *DockerBuildComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DockerBuildComponentResourceModel

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

func (r *DockerBuildComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DockerBuildComponentResourceModel

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

	configRequest := &models.ServiceCreateDockerBuildComponentConfigRequest{
		BuildArgs:  []string{},
		Dockerfile: data.Dockerfile.ValueString(),
		SyncOnly:   data.SyncOnly.ValueBool(),
		Target:     "",
		EnvVars:    map[string]string{},
	}
	if data.BasicDeploy != nil {
		configRequest.BasicDeployConfig = &models.ServiceBasicDeployConfigRequest{
			Args:            []string{},
			CPULimit:        "",
			CPURequest:      "",
			EnvVars:         map[string]string{},
			HealthCheckPath: data.BasicDeploy.HealthCheckPath.String(),
			InstanceCount:   data.BasicDeploy.InstanceCount.ValueInt64(),
			ListenPort:      data.BasicDeploy.Port.ValueInt64(),
			MemLimit:        "",
			MemRequest:      "",
		}
	}
	if data.PublicRepo != nil {
		public := data.PublicRepo
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    public.Branch.ValueStringPointer(),
			Directory: public.Directory.ValueStringPointer(),
			Repo:      public.Repo.ValueStringPointer(),
		}
	} else {
		connected := data.ConnectedRepo
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    connected.Branch.ValueString(),
			Directory: connected.Directory.ValueStringPointer(),
			Repo:      connected.Repo.ValueStringPointer(),
		}
	}
	for _, envVar := range data.EnvVar {
		configRequest.EnvVars[envVar.Name.String()] = envVar.Value.String()
		if configRequest.BasicDeployConfig != nil {
			configRequest.BasicDeployConfig.EnvVars[envVar.Name.String()] = envVar.Value.String()
		}
	}
	_, err = r.restClient.CreateDockerBuildComponentConfig(ctx, data.ID.ValueString(), configRequest)
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component config")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "successfully updated component")
}

func (r *DockerBuildComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
