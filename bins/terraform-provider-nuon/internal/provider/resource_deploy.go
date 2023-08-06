package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/powertoolsdev/mono/pkg/api/gqlclient"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DeployResource{}
var _ resource.ResourceWithImportState = &DeployResource{}

func NewDeployResource() resource.Resource {
	return &DeployResource{}
}

// DeployResource defines the resource implementation.
type DeployResource struct {
	client gqlclient.Client
}

// DeployResourceModel describes the resource data model.
type DeployResourceModel struct {
	ID        types.String `tfsdk:"id"`
	BuildID   types.String `tfsdk:"build_id"`
	InstallID types.String `tfsdk:"install_id"`

	// TODO(jm): remove this when we remove `componentID` from the api-gateway.
	ComponentID types.String `tfsdk:"component_id"`
}

func (r *DeployResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deploy"
}

func (r *DeployResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deploy",
		Attributes: map[string]schema.Attribute{
			"build_id": schema.StringAttribute{
				MarkdownDescription: "build ID",
				Optional:            false,
				Required:            true,
			},
			"component_id": schema.StringAttribute{
				MarkdownDescription: "component ID - must match the id of the build - will be deprecated",
				Optional:            false,
				Required:            true,
			},
			"install_id": schema.StringAttribute{
				MarkdownDescription: "install ID to deploy to",
				Optional:            false,
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "id",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DeployResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(gqlclient.Client)
	if !ok {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, fmt.Errorf("error setting client"), "configure resource")
		return
	}

	r.client = client
}

func (r *DeployResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *DeployResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "start deploy")
	buildResp, err := r.client.StartDeploy(ctx, gqlclient.DeployInput{
		BuildId:     data.BuildID.ValueString(),
		ComponentId: data.ComponentID.ValueString(),
		InstallId:   data.InstallID.ValueString(),
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "start deploy")
		return
	}
	data.ID = types.StringValue(buildResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Trace(ctx, "successfully started deploy")
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(gqlclient.StatusProvisioning), string(gqlclient.StatusUnknown)},
		Target:  []string{string(gqlclient.StatusActive)},
		Refresh: func() (interface{}, string, error) {
			tflog.Trace(ctx, "refreshing instance status")
			status, err := r.client.GetInstanceStatus(ctx,
				data.InstallID.ValueString(),
				data.BuildID.ValueString(),
				data.ID.ValueString(),
			)
			if err != nil {
				return nil, string(gqlclient.StatusUnknown), err
			}
			return status, string(status), nil
		},
		Timeout:    time.Minute * 20,
		Delay:      time.Second * 10,
		MinTimeout: 3 * time.Second,
	}
	statusRaw, err := stateConf.WaitForState()
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "poll deploy")
		return
	}

	status, ok := statusRaw.(gqlclient.Status)
	if !ok {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, fmt.Errorf("invalid deploy status %s", status), "poll deploy")
		return
	}
}

func (r *DeployResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *DeployResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildResp, err := r.client.GetDeploy(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get component")
		return
	}
	data.ID = types.StringValue(buildResp.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeployResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *DeployResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeployResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *DeployResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DeployResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
