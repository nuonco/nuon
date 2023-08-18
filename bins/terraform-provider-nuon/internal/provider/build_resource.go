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
	"github.com/powertoolsdev/mono/pkg/deprecated/api/gqlclient"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &BuildResource{}
var _ resource.ResourceWithImportState = &BuildResource{}

func NewBuildResource() resource.Resource {
	return &BuildResource{}
}

// BuildResource defines the resource implementation.
type BuildResource struct {
	baseResource
}

// BuildResourceModel describes the resource data model.
type BuildResourceModel struct {
	ID          types.String `tfsdk:"id"`
	GitRef      types.String `tfsdk:"git_ref"`
	ComponentID types.String `tfsdk:"component_id"`
}

func (r *BuildResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build"
}

func (r *BuildResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Build",
		Attributes: map[string]schema.Attribute{
			"git_ref": schema.StringAttribute{
				MarkdownDescription: "component",
				Optional:            false,
				Required:            true,
			},
			"component_id": schema.StringAttribute{
				MarkdownDescription: "ID of the component to build",
				Optional:            false,
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of build",
				Computed:            true,
				Required:            false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *BuildResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *BuildResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildResp, err := r.client.StartBuild(ctx, gqlclient.BuildInput{
		GitRef:      data.GitRef.ValueString(),
		ComponentId: data.ComponentID.ValueString(),
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create build")
		return
	}
	data.ID = types.StringValue(buildResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	stateConf := &retry.StateChangeConf{
		Pending: []string{string(gqlclient.StatusProvisioning), string(gqlclient.StatusUnknown)},
		Target:  []string{string(gqlclient.StatusActive)},
		Refresh: func() (interface{}, string, error) {
			tflog.Trace(ctx, "refreshing build status")
			status, err := r.client.GetBuildStatus(ctx,
				data.ID.ValueString(),
			)
			if err != nil {
				writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "poll app")
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
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "polling build")
		return
	}

	status, ok := statusRaw.(gqlclient.Status)
	if !ok {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, fmt.Sprintf("build status %s", status))
		return
	}
}

func (r *BuildResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *BuildResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	buildResp, err := r.client.GetBuild(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get build")
		return
	}
	data.ID = types.StringValue(buildResp.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type buildAWSSettings interface {
	GetRole() string
	GetRegion() string
}

func (r *BuildResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *BuildResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BuildResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *BuildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.client.CancelBuild(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "delete build")
		return
	}
	if !deleted {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "cancel build")
		return
	}
}

func (r *BuildResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root("org_id"), req, resp)
}
