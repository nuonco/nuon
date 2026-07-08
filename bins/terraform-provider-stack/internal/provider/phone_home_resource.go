package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	stack "github.com/nuonco/nuon/sdks/stack"
)

const (
	phoneHomeRequestCreate = "Create"
	phoneHomeRequestUpdate = "Update"
	phoneHomeRequestDelete = "Delete"
)

var (
	_ resource.Resource              = (*phoneHomeResource)(nil)
	_ resource.ResourceWithConfigure = (*phoneHomeResource)(nil)
)

// phoneHomeResource is the stack_phone_home resource. It reports the result of
// an install-stack run to the Nuon control plane over the public phone-home
// endpoint, so an install-stacks module can report run status through the
// provider instead of constructing the HTTP request itself. The resource
// lifecycle drives the request_type: Create on first apply, Update when the
// reported outputs change, Delete on destroy.
type phoneHomeResource struct {
	cfg *providerConfig
}

// NewPhoneHomeResource is the resource factory registered with the provider.
func NewPhoneHomeResource() resource.Resource {
	return &phoneHomeResource{}
}

func (r *phoneHomeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_phone_home"
}

func (r *phoneHomeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *phoneHomeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reports the result of an install-stack run to the Nuon control plane. Create/update/destroy map to the phone-home request_type (Create/Update/Delete). Intended for use inside install-stacks modules so run status is reported through the provider rather than a hand-rolled HTTP call.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Synthetic identifier, `install_id:phone_home_id`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"install_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Nuon install ID (URL path).",
			},
			"phone_home_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Per-stack-version identifier from the control plane; acts as the secret for this report.",
			},
			"phone_home_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Target cloud for the report (`aws` or `gcp`). Merged into the payload as `phone_home_type`.",
			},
			"payload": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The phone-home body as a JSON object string (typically `jsonencode({...})`). The provider injects `request_type` and `phone_home_type`; any values for those keys in the payload are overwritten.",
			},
		},
	}
}

func (r *phoneHomeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data phoneHomeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.report(ctx, &data, phoneHomeRequestCreate); err != nil {
		resp.Diagnostics.AddError("phone home failed", err.Error())
		return
	}
	data.ID = types.StringValue(data.InstallID.ValueString() + ":" + data.PhoneHomeID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *phoneHomeResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// Write-only endpoint: there is no GET to refresh against, so state is
	// preserved as-is.
}

func (r *phoneHomeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data phoneHomeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.report(ctx, &data, phoneHomeRequestUpdate); err != nil {
		resp.Diagnostics.AddError("phone home failed", err.Error())
		return
	}
	data.ID = types.StringValue(data.InstallID.ValueString() + ":" + data.PhoneHomeID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *phoneHomeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data phoneHomeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Best-effort: ctl-api short-circuits Delete, and a failed report must not
	// wedge terraform destroy.
	if err := r.report(ctx, &data, phoneHomeRequestDelete); err != nil {
		resp.Diagnostics.AddWarning("phone home delete failed", err.Error())
	}
}

func (r *phoneHomeResource) report(ctx context.Context, data *phoneHomeResourceModel, requestType string) error {
	payload := map[string]any{}
	if raw := data.Payload.ValueString(); raw != "" {
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return fmt.Errorf("payload must be a JSON object: %w", err)
		}
	}
	payload["request_type"] = requestType
	payload["phone_home_type"] = data.PhoneHomeType.ValueString()

	return stack.PhoneHome(ctx, r.cfg.apiURL, data.InstallID.ValueString(), data.PhoneHomeID.ValueString(), payload)
}
