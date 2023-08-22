package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvVar struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func envVarSharedBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "env var key",
					Required:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Value - can be interpolated from any nuon value",
					Required:            true,
				},
			},
		},
	}
}

type HelmValue struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func helmValueSharedBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "helm values to set, such as `env.secret` or `server.container.image`",
					Required:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Value - can be interpolated from any nuon value",
					Required:            true,
				},
			},
		},
	}
}

type TerraformVariable struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func terraformVariableSharedBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: "Terraform variable to get set. By default is rendered into a tfvars file in the run",
					Required:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "Value - can be interpolated from any nuon value",
					Required:            true,
				},
			},
		},
	}
}
