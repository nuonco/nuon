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
var _ resource.Resource = &InstallResource{}
var _ resource.ResourceWithImportState = &InstallResource{}

func NewInstallResource() resource.Resource {
	return &InstallResource{}
}

// InstallResource defines the resource implementation.
type InstallResource struct {
	client gqlclient.Client
}

// InstallResourceModel describes the resource data model.
type InstallResourceModel struct {
	Name       types.String `tfsdk:"name"`
	AppID      types.String `tfsdk:"app_id"`
	Region     types.String `tfsdk:"region"`
	IAMRoleARN types.String `tfsdk:"iam_role_arn"`

	// computed
	ID types.String `tfsdk:"id"`
}

func (r *InstallResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_install"
}

func (r *InstallResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Install",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the application.",
				Optional:            false,
				Required:            true,
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "ID of the app this install belongs too.",
				Optional:            false,
				Required:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "AWS region",
				Optional:            false,
				Required:            true,
			},
			"iam_role_arn": schema.StringAttribute{
				MarkdownDescription: "ARN of the role to use for provisioning.",
				Optional:            false,
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the install",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *InstallResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InstallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *InstallResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "creating install")
	region, err := stringToAPIRegion(data.Region.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "create install")
		return
	}

	installResp, err := r.client.UpsertInstall(ctx, gqlclient.InstallInput{
		Name:  data.Name.ValueString(),
		AppId: data.AppID.ValueString(),
		AwsSettings: &gqlclient.AWSSettingsInput{
			Region: region,
			Role:   data.IAMRoleARN.ValueString(),
		},
	})
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "create install")
		return
	}
	data.ID = types.StringValue(installResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Trace(ctx, "successfully created app")

	stateConf := &retry.StateChangeConf{
		Pending: []string{string(gqlclient.StatusProvisioning), string(gqlclient.StatusUnknown)},
		Target:  []string{string(gqlclient.StatusActive)},
		Refresh: func() (interface{}, string, error) {
			tflog.Trace(ctx, "refreshing install status")
			status, err := r.client.GetInstallStatus(ctx,
				installResp.App.Org.Id,
				data.AppID.ValueString(),
				data.ID.ValueString(),
			)
			if err != nil {
				writeDiagnosticsErr(ctx, resp.Diagnostics, err, "poll status")
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
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "get install")
		return
	}

	status, ok := statusRaw.(gqlclient.Status)
	if !ok {
		writeDiagnosticsErr(ctx, resp.Diagnostics, fmt.Errorf("invalid install %s", status), "create install")
		return
	}
}

func (r *InstallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *InstallResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	installResp, err := r.client.GetInstall(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "get install")
		return
	}
	data.Name = types.StringValue(installResp.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type installAWSSettings interface {
	GetRole() string
	GetRegion() string
}

func (r *InstallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *InstallResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	installResp, err := r.client.UpsertInstall(ctx, gqlclient.InstallInput{
		Id:   data.ID.ValueString(),
		Name: data.Name.ValueString(),
	})
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "update install")
		return
	}

	if installResp.GetSettings().(installAWSSettings).GetRole() != data.IAMRoleARN.ValueString() {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "IAM Role ARN changed")
		return
	}

	if installResp.GetSettings().(installAWSSettings).GetRegion() != data.Region.ValueString() {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "AWS Region changed")
		return
	}

	data.ID = types.StringValue(installResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *InstallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.client.DeleteInstall(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "delete install")
		return
	}
	if !deleted {
		writeDiagnosticsErr(ctx, resp.Diagnostics, err, "delete install")
		return
	}
}

func (r *InstallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root("org_id"), req, resp)
}
