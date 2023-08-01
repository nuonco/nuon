package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/powertoolsdev/mono/pkg/api/gqlclient"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ContainerImageComponentResource{}
var _ resource.ResourceWithImportState = &ContainerImageComponentResource{}

func NewContainerImageComponentResource() resource.Resource {
	return &ContainerImageComponentResource{}
}

// ContainerImageComponentResource defines the resource implementation.
type ContainerImageComponentResource struct {
	client gqlclient.Client
}

type AwsEcr struct {
	Region     types.String `tfsdk:"region"`
	Tag        types.String `tfsdk:"tag"`
	ImageURL   types.String `tfsdk:"image_url"`
	IAMRoleARN types.String `tfsdk:"iam_role_arn"`
}

// ContainerImageComponentResourceModel describes the resource data model.
type ContainerImageComponentResourceModel struct {
	ID types.String `tfsdk:"id"`

	Name     types.String `tfsdk:"name"`
	AppID    types.String `tfsdk:"app_id"`
	SyncOnly types.Bool   `tfsdk:"sync_only"`

	BasicDeploy *BasicDeploy `tfsdk:"basic_deploy"`

	AwsEcr *AwsEcr `tfsdk:"aws_ecr"`
	Public struct {
		ImageURL types.String `tfsdk:"image_url"`
		Tag      types.String `tfsdk:"tag"`
	} `tfsdk:"public"`

	EnvVar []EnvVar `tfsdk:"env_var"`
}

func (r *ContainerImageComponentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_image_component"
}

func (r *ContainerImageComponentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContainerImageComponentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Container images are used to connect any Docker, ECR or OCI compatible image to your app.",
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
			"sync_only": schema.BoolAttribute{
				MarkdownDescription: "Set to true to only use this image for syncing (ie: no deployment).",
				Optional:            true,
				Required:            false,
			},

			// public
			"public": schema.SingleNestedAttribute{
				Description: "any public, Docker or oci image",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"image_url": schema.StringAttribute{
						MarkdownDescription: "full image url, or docker hub alias (kennethreitz/httpbin)",
						Required:            true,
					},
					"tag": schema.StringAttribute{
						MarkdownDescription: "tag",
						Required:            true,
					},
				},
			},
			"aws_ecr": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "any image stored in ECR, with an IAM role that your org can assume.",
				Attributes: map[string]schema.Attribute{
					"region": schema.StringAttribute{
						MarkdownDescription: "ECR region",
						Required:            true,
					},
					"tag": schema.StringAttribute{
						MarkdownDescription: "tag",
						Required:            true,
					},
					"image_url": schema.StringAttribute{
						MarkdownDescription: "image_url",
						Required:            true,
					},
					"iam_role_arn": schema.StringAttribute{
						MarkdownDescription: "iam_role_arn",
						Required:            true,
					},
				},
			},
			"basic_deploy": basicDeployAttribute(),
		},
		Blocks: map[string]schema.Block{
			"env_var": envVarSharedBlock(),
		},
	}
}

func (r *ContainerImageComponentResource) getConfigInput(data *ContainerImageComponentResourceModel) (*gqlclient.ComponentConfigInput, error) {
	envVars := make([]*gqlclient.KeyValuePairInput, 0)
	for _, envVar := range data.EnvVar {
		envVars = append(envVars, envVar.toKeyValueInput())
	}

	cfg := &gqlclient.ComponentConfigInput{
		BuildConfig:  &gqlclient.BuildConfigInput{},
		DeployConfig: &gqlclient.DeployConfigInput{},
	}
	if data.SyncOnly.ValueBool() {
		cfg.DeployConfig.Noop = true
	} else {
		cfg.DeployConfig = data.BasicDeploy.toDeployConfigInput()
		cfg.DeployConfig.BasicDeployConfig.EnvVars = envVars
	}

	if !data.Public.ImageURL.IsNull() {
		cfg.BuildConfig.ExternalImageConfig = &gqlclient.ExternalImageInput{
			OciImageUrl: data.Public.ImageURL.ValueString(),
			Tag:         data.Public.Tag.ValueString(),
		}
	} else {
		region, err := stringToAPIRegion(data.AwsEcr.Region.ValueString())
		if err != nil {
			return cfg, fmt.Errorf("invalid region: %w", err)
		}

		cfg.BuildConfig.ExternalImageConfig = &gqlclient.ExternalImageInput{
			OciImageUrl: data.AwsEcr.ImageURL.ValueString(),
			Tag:         data.AwsEcr.Tag.ValueString(),
			AuthConfig: &gqlclient.AWSAuthConfigInput{
				Role:   data.AwsEcr.IAMRoleARN.ValueString(),
				Region: region,
			},
		}
	}

	return cfg, nil
}

func (r *ContainerImageComponentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ContainerImageComponentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfgInput, err := r.getConfigInput(data)
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "map component configuration")
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

	tflog.Trace(ctx, "got ID -- "+compResp.Id)
	data.ID = types.StringValue(compResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Trace(ctx, "successfully created component")
}

func (r *ContainerImageComponentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ContainerImageComponentResourceModel

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

func (r *ContainerImageComponentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ContainerImageComponentResourceModel

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

func (r *ContainerImageComponentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ContainerImageComponentResourceModel

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
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "update component")
		return
	}

	data.Name = types.StringValue(installResp.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContainerImageComponentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
