package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	stack "github.com/nuonco/nuon/sdks/stack"
)

var (
	_ datasource.DataSource              = (*stackDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*stackDataSource)(nil)
)

// stackDataSource is the stack_config data source. It reads an install stack's
// rendered configuration from the Nuon control plane (keyed by phone_home_id)
// without provisioning anything, so an install-stacks module can consume the
// config directly instead of receiving it as tfvars.
type stackDataSource struct {
	cfg *providerConfig
}

// NewStackDataSource is the data source factory registered with the provider.
func NewStackDataSource() datasource.DataSource {
	return &stackDataSource{}
}

func (d *stackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (d *stackDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(*providerConfig)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data", fmt.Sprintf("expected *providerConfig, got %T", req.ProviderData))
		return
	}
	d.cfg = cfg
}

func (d *stackDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	roleAttrs := map[string]schema.Attribute{
		"permissions":     schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "IAM permissions bound to the role's service account."},
		"predefined_role": schema.StringAttribute{Computed: true, MarkdownDescription: "Predefined role bound to the service account, if any."},
		"enabled":         schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the role should be created."},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Nuon install stack's rendered configuration (runner, permissions, inputs, secrets) from the control plane. Intended for use inside install-stacks modules so the config is read from the API rather than passed in as tfvars.",
		Attributes: map[string]schema.Attribute{
			"phone_home_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Per-stack-version identifier from the Nuon control plane; acts as the secret for this read.",
			},

			"install_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Nuon install ID."},
			"org_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Nuon organization ID."},
			"app_id":         schema.StringAttribute{Computed: true, MarkdownDescription: "Nuon application ID."},
			"cloud":          schema.StringAttribute{Computed: true, MarkdownDescription: "Target cloud (aws or gcp)."},
			"runner_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Runner ID for this install."},
			"runner_api_url": schema.StringAttribute{Computed: true, MarkdownDescription: "Runner API URL the runner reports to."},
			"phone_home_url": schema.StringAttribute{Computed: true, MarkdownDescription: "Phone-home URL the module reports run completion to."},

			"install_inputs": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Customer install-input values.",
			},
			"auto_generate_secrets": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Names of secrets the stack should auto-generate.",
			},
			"secrets": schema.MapNestedAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Customer-supplied secrets, keyed by name.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Secret description."},
						"required":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the secret is required."},
						"value":       schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Secret value."},
					},
				},
			},

			"gcp": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "GCP-specific configuration. Present when cloud is gcp.",
				Attributes: map[string]schema.Attribute{
					"runner_init_script_url": schema.StringAttribute{Computed: true, MarkdownDescription: "Runner bootstrap script URL."},
					"runner_api_token":       schema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Runner API token."},

					"provision_permissions":       schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Provision service-account permissions."},
					"provision_predefined_role":   schema.StringAttribute{Computed: true, MarkdownDescription: "Provision predefined role, if any."},
					"maintenance_permissions":     schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Maintenance service-account permissions."},
					"maintenance_predefined_role": schema.StringAttribute{Computed: true, MarkdownDescription: "Maintenance predefined role, if any."},
					"deprovision_permissions":     schema.ListAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Deprovision service-account permissions."},
					"deprovision_predefined_role": schema.StringAttribute{Computed: true, MarkdownDescription: "Deprovision predefined role, if any."},

					"break_glass_roles": schema.MapNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Break-glass roles, keyed by name.",
						NestedObject:        schema.NestedAttributeObject{Attributes: roleAttrs},
					},
					"custom_roles": schema.MapNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Customer-defined roles, keyed by name.",
						NestedObject:        schema.NestedAttributeObject{Attributes: roleAttrs},
					},
				},
			},
		},
	}
}

func (d *stackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data stackDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := strings.TrimRight(d.cfg.apiURL, "/") + "/v1/stack-runs/" + data.PhoneHomeID.ValueString()
	cfg, err := stack.FetchConfig(ctx, url)
	if err != nil {
		resp.Diagnostics.AddError("fetch stack config failed", err.Error())
		return
	}

	flattenConfig(&data, cfg)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
