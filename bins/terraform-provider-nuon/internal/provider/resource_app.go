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
var _ resource.Resource = &AppResource{}
var _ resource.ResourceWithImportState = &AppResource{}

func NewAppResource() resource.Resource {
	return &AppResource{}
}

// AppResource defines the resource implementation.
type AppResource struct {
	client gqlclient.Client
}

// AppResourceModel describes the resource data model.
type AppResourceModel struct {
	Name  types.String `tfsdk:"name"`
	OrgID types.String `tfsdk:"org_id"`
	Id    types.String `tfsdk:"id"`
}

func (r *AppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *AppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Application",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the application.",
				Optional:            false,
				Required:            true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "ID of the org this app belongs too.",
				Optional:            false,
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the app",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *AppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AppResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "creating app")
	appResp, err := r.client.UpsertApp(ctx, gqlclient.AppInput{
		Name:  data.Name.ValueString(),
		OrgId: data.OrgID.ValueString(),
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "create app")
		return
	}
	data.Name = types.StringValue(appResp.Name)
	data.Id = types.StringValue(appResp.Id)
	data.OrgID = types.StringValue(data.OrgID.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Trace(ctx, "successfully created app")
}

func (r *AppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AppResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appResp, err := r.client.GetApp(ctx, data.Id.ValueString())
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "get app")
		return
	}

	data.Name = types.StringValue(appResp.Name)
	data.Id = types.StringValue(appResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AppResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appResp, err := r.client.UpsertApp(ctx, gqlclient.AppInput{
		Id:   data.Id.ValueString(),
		Name: data.Name.ValueString(),
	})
	if err != nil {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "upsert app")
		return
	}

	data.Name = types.StringValue(appResp.Name)
	data.Id = types.StringValue(appResp.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleted, err := r.client.DeleteApp(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete app",
			fmt.Sprintf("Please make sure your app_id (%s) is correct, and that the auth token has permissions for this org. ", data.Id.String()),
		)
		return
	}
	if !deleted {
		writeDiagnosticsErr(ctx, &resp.Diagnostics, err, "delete app")
		return
	}

}

func (r *AppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root("org_id"), req, resp)
}
