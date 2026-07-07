package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	stack "github.com/nuonco/nuon/sdks/stack"
)

var (
	_ resource.Resource                = (*stackResource)(nil)
	_ resource.ResourceWithConfigure   = (*stackResource)(nil)
	_ resource.ResourceWithImportState = (*stackResource)(nil)
)

// stackResource is the nuon_stack resource.
type stackResource struct {
	cfg *providerConfig
}

// NewStackResource is the resource factory registered with the provider.
func NewStackResource() resource.Resource {
	return &stackResource{}
}

func (r *stackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (r *stackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(*providerConfig)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data", fmt.Sprintf("expected *providerConfig, got %T", req.ProviderData))
		return
	}
	r.cfg = cfg
}

func (r *stackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Nuon install stack, provisioned locally via the stack SDK against the customer's cloud account.",
		Attributes: map[string]schema.Attribute{
			"phone_home_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Per-stack-version identifier from the Nuon control plane. Changing it replaces the resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"terraform_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Terraform version to run (hc-install pin). Empty resolves the latest stable release.",
			},
			"module_ref": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Override for the install-stacks module ref/URL. Empty uses the default main archive.",
			},
			"inputs": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Customer install-input values. Keys must be declared by the app config.",
			},
			"secrets": schema.MapAttribute{
				Optional:            true,
				Sensitive:           true,
				ElementType:         types.StringType,
				MarkdownDescription: "Customer secret values (write-only). Keys must be declared by the app config; auto-generated secrets are rejected.",
			},
			"install_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Nuon install ID resolved by the control plane. Also the resource ID.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"outputs": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Flattened install-stack outputs (ARNs, IDs, resolved inputs).",
			},
		},
		Blocks: map[string]schema.Block{
			"aws": schema.SingleNestedBlock{
				MarkdownDescription: "AWS target configuration. Present for AWS stacks.",
				Attributes: map[string]schema.Attribute{
					"region": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "AWS region. Resolved from the control plane when omitted.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"account_id": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "AWS account ID. Validated against the resolved account; resolved when omitted.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"state_bucket": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Existing S3 bucket in the target account for remote Terraform state.",
					},
					"state_key": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "S3 state object key. Defaults to nuon/<install_id>/terraform.tfstate.",
						PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
					},
					"dynamodb_table": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Optional DynamoDB table for state locking (or rely on S3 native locking).",
					},
				},
			},
		},
	}
}

func (r *stackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data stackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.AWS == nil {
		resp.Diagnostics.AddError("missing cloud block", "an aws {} block is required")
		return
	}

	opts := data.urlOptions(ctx, r.cfg.apiURL, &resp.Diagnostics)
	opts.Kind = stack.KindProvision
	if resp.Diagnostics.HasError() {
		return
	}

	inst, err := stack.FromURL(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("create stack run failed", err.Error())
		return
	}
	defer inst.Close(ctx)

	prepared := inst.PreparedConfig()
	out, err := inst.Provision(ctx)
	if err != nil {
		resp.Diagnostics.AddError("provision failed", err.Error())
		return
	}

	r.applyResult(ctx, &data, prepared, out, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *stackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data stackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region := ""
	if data.AWS != nil {
		region = data.AWS.Region.ValueString()
	}
	inst, err := stack.New(ctx, stack.Options{
		InstallID: data.InstallID.ValueString(),
		Cloud:     stack.CloudAWS,
		AWSRegion: region,
		Backend:   r.backendFromModel(&data),
	})
	if err != nil {
		resp.Diagnostics.AddError("read stack failed", err.Error())
		return
	}
	defer inst.Close(ctx)

	out, err := inst.Status(ctx)
	if err != nil {
		// Treat a failed status read as drift rather than a hard error would be
		// ideal, but distinguishing "gone" from "transient" needs richer SDK
		// signals; surface the error for now.
		resp.Diagnostics.AddError("read stack status failed", err.Error())
		return
	}

	// Secrets are write-only: never reconcile them from Status.
	r.applyOutputs(ctx, &data, out, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *stackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data stackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := data.urlOptions(ctx, r.cfg.apiURL, &resp.Diagnostics)
	opts.Kind = stack.KindReprovision
	if resp.Diagnostics.HasError() {
		return
	}

	inst, err := stack.FromURL(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("create reprovision run failed", err.Error())
		return
	}
	defer inst.Close(ctx)

	prepared := inst.PreparedConfig()
	out, err := inst.Reprovision(ctx)
	if err != nil {
		resp.Diagnostics.AddError("reprovision failed", err.Error())
		return
	}

	r.applyResult(ctx, &data, prepared, out, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *stackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data stackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := data.urlOptions(ctx, r.cfg.apiURL, &resp.Diagnostics)
	opts.Kind = stack.KindDeprovision
	if resp.Diagnostics.HasError() {
		return
	}

	inst, err := stack.FromURL(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("create deprovision run failed", err.Error())
		return
	}
	defer inst.Close(ctx)

	if err := inst.Deprovision(ctx); err != nil {
		resp.Diagnostics.AddError("deprovision failed", err.Error())
		return
	}
}

// ImportState accepts "install_id:region" so Read can reconstruct the SDK
// client. phone_home_id must be re-supplied in config after import.
func (r *stackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	installID, region, ok := strings.Cut(req.ID, ":")
	if !ok || installID == "" || region == "" {
		resp.Diagnostics.AddError("invalid import ID", "expected format install_id:region")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("install_id"), installID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aws").AtName("region"), region)...)
}

// applyResult writes the resolved install ID, AWS block defaults, and outputs
// back onto the model after a provision/reprovision.
func (r *stackResource) applyResult(ctx context.Context, data *stackResourceModel, prepared *stack.Config, out *stack.Outputs, diags *diag.Diagnostics) {
	installID := data.InstallID.ValueString()
	if prepared != nil && prepared.InstallID != "" {
		installID = prepared.InstallID
	}
	data.InstallID = types.StringValue(installID)

	if data.AWS != nil && out != nil && out.AWS != nil {
		if data.AWS.Region.IsNull() || data.AWS.Region.IsUnknown() || data.AWS.Region.ValueString() == "" {
			data.AWS.Region = types.StringValue(out.AWS.Region)
		}
		if data.AWS.AccountID.IsNull() || data.AWS.AccountID.IsUnknown() || data.AWS.AccountID.ValueString() == "" {
			data.AWS.AccountID = types.StringValue(out.AWS.AccountID)
		}
		if data.AWS.StateKey.IsNull() || data.AWS.StateKey.IsUnknown() || data.AWS.StateKey.ValueString() == "" {
			data.AWS.StateKey = types.StringValue(fmt.Sprintf("nuon/%s/terraform.tfstate", installID))
		}
	}

	r.applyOutputs(ctx, data, out, diags)
}

func (r *stackResource) applyOutputs(ctx context.Context, data *stackResourceModel, out *stack.Outputs, diags *diag.Diagnostics) {
	m, d := types.MapValueFrom(ctx, types.StringType, flattenOutputs(out))
	diags.Append(d...)
	data.Outputs = m
}

func (r *stackResource) backendFromModel(data *stackResourceModel) stack.TerraformBackend {
	if data.AWS == nil {
		return stack.TerraformBackend{}
	}
	return stack.TerraformBackend{
		Bucket:        data.AWS.StateBucket.ValueString(),
		Key:           data.AWS.StateKey.ValueString(),
		Region:        data.AWS.Region.ValueString(),
		DynamoDBTable: data.AWS.DynamoDBTable.ValueString(),
	}
}
