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
	"github.com/powertoolsdev/mono/pkg/deprecated/api/gqlclient"
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

func (r *TerraformModuleComponentResource) getConfigInput(data *TerraformModuleComponentResourceModel) (*gqlclient.ComponentConfigInput, error) {
	tfVars := make([]*gqlclient.KeyValuePairInput, 0)
	for _, tfVar := range data.Var {
		tfVars = append(tfVars, tfVar.toKeyValueInput())
	}

	cfg := &gqlclient.ComponentConfigInput{
		BuildConfig: &gqlclient.BuildConfigInput{
			TerraformBuildConfig: &gqlclient.TerraformBuildInput{},
		},
		DeployConfig: &gqlclient.DeployConfigInput{
			TerraformDeployConfig: &gqlclient.TerraformDeployConfigInput{
				TerraformVersion: gqlclient.TerraformVersionTerraformVersionLatest,
				Vars:             tfVars,
			},
		},
	}

	if data.PublicRepo != nil {
		cfg.BuildConfig.TerraformBuildConfig.VcsConfig = data.PublicRepo.getVCSConfig()
	} else {
		cfg.BuildConfig.TerraformBuildConfig.VcsConfig = data.ConnectedRepo.getVCSConfig()
	}
	return cfg, nil
}

func (r *TerraformModuleComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *TerraformModuleComponentResourceModel
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

	tflog.Trace(ctx, "successfully created component")
}

func (r *TerraformModuleComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *TerraformModuleComponentResourceModel

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

func (r *TerraformModuleComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *TerraformModuleComponentResourceModel

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

func (r *TerraformModuleComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *TerraformModuleComponentResourceModel

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

func (r *TerraformModuleComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
