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
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &InstallResource{}
var _ resource.ResourceWithImportState = &InstallResource{}

func NewInstallResource() resource.Resource {
	return &InstallResource{}
}

// InstallResource defines the resource implementation.
type InstallResource struct {
	baseResource
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "AWS region",
				Optional:            false,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"iam_role_arn": schema.StringAttribute{
				MarkdownDescription: "ARN of the role to use for provisioning.",
				Optional:            false,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *InstallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *InstallResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "creating install")

	installResp, err := r.restClient.CreateInstall(ctx, data.AppID.ValueString(), &models.ServiceCreateInstallRequest{
		Name: data.Name.ValueStringPointer(),
		AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
			Region:     data.Region.ValueString(),
			IamRoleArn: data.IAMRoleARN.ValueStringPointer(),
		},
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create install")
		return
	}
	data.ID = types.StringValue(installResp.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Trace(ctx, "successfully created install")

	stateConf := &retry.StateChangeConf{
		Pending: []string{statusQueued, statusProvisioning},
		Target:  []string{statusActive},
		Refresh: func() (interface{}, string, error) {
			tflog.Trace(ctx, "refreshing install status")
			install, err := r.restClient.GetInstall(ctx, installResp.ID)
			if err != nil {
				writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "poll status")
				return nil, "unknown", err
			}
			return install.Status, string(install.Status), nil
		},
		Timeout:    time.Minute * 20,
		Delay:      time.Second * 10,
		MinTimeout: 3 * time.Second,
	}
	statusRaw, err := stateConf.WaitForState()
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get install")
		return
	}

	status, ok := statusRaw.(string)
	if !ok {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, fmt.Errorf("invalid install %s", status), "create install")
		return
	}
}

func (r *InstallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *InstallResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	installResp, err := r.restClient.GetInstall(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get install")
		return
	}
	data.Name = types.StringValue(installResp.Name)
	data.AppID = types.StringValue(installResp.AppID)
	data.IAMRoleARN = types.StringValue(installResp.AwsAccount.IamRoleArn)
	data.Region = types.StringValue(installResp.AwsAccount.Region)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *InstallResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	installResp, err := r.restClient.UpdateInstall(ctx, data.ID.ValueString(), &models.ServiceUpdateInstallRequest{
		Name: data.Name.ValueString(),
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "update install")
		return
	}
	data.ID = types.StringValue(installResp.ID)
	data.Name = types.StringValue(installResp.Name)

	// TODO: The SDK doesn't return these values.
	// These can't be updated anyway, so it's not a blocker,
	// but it would be ideal to use the API as the source of truth.
	// data.AppID = types.StringValue(installResp.AppID)
	// data.IAMRoleARN = types.StringValue(installResp.AwsAccount.IamRoleArn)
	// data.Region = types.StringValue(installResp.AwsAccount.Region)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *InstallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.restClient.DeleteInstall(ctx, data.ID.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "delete install")
		return
	}
	if !deleted {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "delete install")
		return
	}

	data.ID = types.StringValue(data.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleteQueued, statusDeprovisioning},
		Target:  []string{""},
		Refresh: func() (interface{}, string, error) {
			tflog.Trace(ctx, "refreshing install status")
			install, err := r.restClient.GetInstall(ctx, data.ID.ValueString())
			if err != nil {
				return "", "", nil
			} else {
				return install.Status, install.Status, nil
			}
		},
		Timeout:    time.Minute * 20,
		Delay:      time.Second * 10,
		MinTimeout: 3 * time.Second,
	}
	statusRaw, err := stateConf.WaitForState()
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get install")
		return
	}

	status, ok := statusRaw.(string)
	if !ok {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, fmt.Errorf("invalid install %s", status), "create install")
	}
}

func (r *InstallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
