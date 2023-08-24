package provider

import (
	"context"

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
var _ resource.Resource = &TerraformModuleComponentResource{}
var _ resource.ResourceWithImportState = &TerraformModuleComponentResource{}

func NewTerraformModuleComponentResource() resource.Resource {
	return &TerraformModuleComponentResource{}
}

// TerraformModuleComponentResource defines the resource implementation.
type TerraformModuleComponentResource struct {
	baseResource
}

// TerraformModuleComponentResourceModel describes the resource data model.
type TerraformModuleComponentResourceModel struct {
	ID types.String `tfsdk:"id"`

	Name             types.String        `tfsdk:"name"`
	AppID            types.String        `tfsdk:"app_id"`
	TerraformVersion types.String        `tfsdk:"terraform_version"`
	PublicRepo       *PublicRepo         `tfsdk:"public_repo"`
	ConnectedRepo    *ConnectedRepo      `tfsdk:"connected_repo"`
	Var              []TerraformVariable `tfsdk:"var"`
}

func (r *TerraformModuleComponentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_terraform_module_component"
}

func (r *TerraformModuleComponentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deploy a terraform module in a public connected repo.",
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
			"terraform_version": schema.StringAttribute{
				MarkdownDescription: "Terraform version to run as.",
				Optional:            true,
				Default:             stringdefault.StaticString("1.5.3"),
				Computed:            true,
			},
			"public_repo":    publicRepoAttribute(),
			"connected_repo": connectedRepoAttribute(),
		},
		Blocks: map[string]schema.Block{
			"var": terraformVariableSharedBlock(),
		},
	}
}

func (r *TerraformModuleComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TerraformModuleComponentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "creating component")

	compResp, err := r.restClient.CreateComponent(ctx, data.AppID.ValueString(), &models.ServiceCreateComponentRequest{
		Name: data.Name.ValueStringPointer(),
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component")
		return
	}
	tflog.Trace(ctx, "got ID -- "+compResp.ID)
	data.ID = types.StringValue(compResp.ID)

	configRequest := &models.ServiceCreateTerraformModuleComponentConfigRequest{
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Variables:                map[string]string{},
		Version:                  data.TerraformVersion.ValueString(),
	}
	for _, value := range data.Var {
		configRequest.Variables[value.Name.String()] = value.Value.String()
	}
	if data.PublicRepo != nil {
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    data.PublicRepo.Branch.ValueStringPointer(),
			Directory: data.PublicRepo.Directory.ValueStringPointer(),
			Repo:      data.PublicRepo.Repo.ValueStringPointer(),
		}
	} else {
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    data.ConnectedRepo.Branch.ValueString(),
			Directory: data.ConnectedRepo.Directory.ValueStringPointer(),
			Repo:      data.ConnectedRepo.Repo.ValueStringPointer(),
		}
	}
	_, err = r.restClient.CreateTerraformModuleComponentConfig(ctx, compResp.ID, configRequest)
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component config")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "successfully created component")
}

func (r *TerraformModuleComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TerraformModuleComponentResourceModel

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
	data.TerraformVersion = types.StringValue(configResp.TerraformModule.Version)
	for key, val := range configResp.TerraformModule.Variables {
		data.Var = append(data.Var, TerraformVariable{
			Name:  types.StringValue(key),
			Value: types.StringValue(val),
		})
	}
	if configResp.Helm.PublicGitVcsConfig != nil {
		data.PublicRepo.Branch = types.StringValue(configResp.Helm.PublicGitVcsConfig.Branch)
		data.PublicRepo.Directory = types.StringValue(configResp.Helm.PublicGitVcsConfig.Directory)
		data.PublicRepo.Repo = types.StringValue(configResp.Helm.PublicGitVcsConfig.Repo)
	} else {
		data.ConnectedRepo.Branch = types.StringValue(configResp.Helm.ConnectedGithubVcsConfig.Branch)
		data.ConnectedRepo.Directory = types.StringValue(configResp.Helm.ConnectedGithubVcsConfig.Directory)
		data.ConnectedRepo.Repo = types.StringValue(configResp.Helm.ConnectedGithubVcsConfig.Repo)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Trace(ctx, "successfully read component")
}

func (r *TerraformModuleComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TerraformModuleComponentResourceModel

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
}

func (r *TerraformModuleComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *TerraformModuleComponentResourceModel

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

	configRequest := &models.ServiceCreateTerraformModuleComponentConfigRequest{
		ConnectedGithubVcsConfig: nil,
		PublicGitVcsConfig:       nil,
		Variables:                map[string]string{},
		Version:                  data.TerraformVersion.ValueString(),
	}
	for _, value := range data.Var {
		configRequest.Variables[value.Name.String()] = value.Value.String()
	}
	if data.PublicRepo != nil {
		configRequest.PublicGitVcsConfig = &models.ServicePublicGitVCSConfigRequest{
			Branch:    data.PublicRepo.Branch.ValueStringPointer(),
			Directory: data.PublicRepo.Directory.ValueStringPointer(),
			Repo:      data.PublicRepo.Repo.ValueStringPointer(),
		}
	} else {
		configRequest.ConnectedGithubVcsConfig = &models.ServiceConnectedGithubVCSConfigRequest{
			Branch:    data.ConnectedRepo.Branch.ValueString(),
			Directory: data.ConnectedRepo.Directory.ValueStringPointer(),
			Repo:      data.ConnectedRepo.Repo.ValueStringPointer(),
		}
	}
	_, err = r.restClient.CreateTerraformModuleComponentConfig(ctx, compResp.ID, configRequest)
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create component config")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TerraformModuleComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
