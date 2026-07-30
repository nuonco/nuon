package componenthealth

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func managed(address, typ, name, providerName string, values map[string]any) *tfjson.StateResource {
	return &tfjson.StateResource{
		Address:         address,
		Mode:            tfjson.ManagedResourceMode,
		Type:            typ,
		Name:            name,
		ProviderName:    providerName,
		AttributeValues: values,
	}
}

func TestTerraformResourceRows(t *testing.T) {
	tests := []struct {
		name  string
		state *tfjson.State
		want  []*models.ServiceComponentHealthResource
	}{
		{
			name:  "nil_state",
			state: nil,
		},
		{
			name:  "no_values",
			state: &tfjson.State{},
		},
		{
			name:  "empty_state",
			state: &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{}}},
		},
		{
			name: "aws_resource",
			state: &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{
				Resources: []*tfjson.StateResource{
					managed("aws_db_instance.main", "aws_db_instance", "main",
						"registry.terraform.io/hashicorp/aws",
						map[string]any{"id": "db-abc123", "arn": "arn:aws:rds:us-west-2:1234:db:main"}),
				},
			}}},
			want: []*models.ServiceComponentHealthResource{{
				Provider: providerAWS,
				Kind:     "aws_db_instance",
				Name:     "db-abc123",
				Health:   healthUnknown,
				Details:  `{"terraform":{"address":"aws_db_instance.main","id":"db-abc123","region":"us-west-2"}}`,
			}},
		},
		{
			name: "skips_data_sources",
			state: &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{
				Resources: []*tfjson.StateResource{
					{
						Address:         "data.aws_ami.ubuntu",
						Mode:            tfjson.DataResourceMode,
						Type:            "aws_ami",
						Name:            "ubuntu",
						ProviderName:    "registry.terraform.io/hashicorp/aws",
						AttributeValues: map[string]any{"id": "ami-123"},
					},
					// no mode recorded: the address is the only marker
					{
						Address:         "data.aws_caller_identity.current",
						Type:            "aws_caller_identity",
						Name:            "current",
						ProviderName:    "registry.terraform.io/hashicorp/aws",
						AttributeValues: map[string]any{"id": "1234"},
					},
					managed("aws_s3_bucket.assets", "aws_s3_bucket", "assets",
						"registry.terraform.io/hashicorp/aws",
						map[string]any{"id": "assets-bucket", "region": "eu-west-1"}),
				},
			}}},
			want: []*models.ServiceComponentHealthResource{{
				Provider: providerAWS,
				Kind:     "aws_s3_bucket",
				Name:     "assets-bucket",
				Health:   healthUnknown,
				Details:  `{"terraform":{"address":"aws_s3_bucket.assets","id":"assets-bucket","region":"eu-west-1"}}`,
			}},
		},
		{
			name: "skips_pseudo_resources",
			state: &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{
				Resources: []*tfjson.StateResource{
					managed("null_resource.wait", "null_resource", "wait",
						"registry.terraform.io/hashicorp/null", nil),
					managed("random_password.db", "random_password", "db",
						"registry.terraform.io/hashicorp/random", nil),
					managed("local_file.cfg", "local_file", "cfg",
						"registry.terraform.io/hashicorp/local", nil),
					managed("tls_private_key.k", "tls_private_key", "k",
						"registry.terraform.io/hashicorp/tls", nil),
					managed("helm_release.app", "helm_release", "app",
						"registry.terraform.io/hashicorp/helm", map[string]any{"name": "app"}),
					managed("kubernetes_namespace.ns", "kubernetes_namespace", "ns",
						"registry.terraform.io/hashicorp/kubernetes", map[string]any{"id": "ns"}),
				},
			}}},
		},
		{
			name: "walks_child_modules",
			state: &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{
				Resources: []*tfjson.StateResource{
					managed("aws_vpc.main", "aws_vpc", "main",
						"registry.terraform.io/hashicorp/aws", map[string]any{"id": "vpc-1"}),
				},
				ChildModules: []*tfjson.StateModule{
					{
						Address: "module.db",
						Resources: []*tfjson.StateResource{
							managed("module.db.google_sql_database_instance.pg", "google_sql_database_instance", "pg",
								"registry.terraform.io/hashicorp/google",
								map[string]any{"name": "pg-prod", "region": "us-central1"}),
							{
								Address:         "module.db.data.google_project.this",
								Mode:            tfjson.DataResourceMode,
								Type:            "google_project",
								Name:            "this",
								ProviderName:    "registry.terraform.io/hashicorp/google",
								AttributeValues: map[string]any{"id": "proj"},
							},
						},
						ChildModules: []*tfjson.StateModule{
							{
								Address: "module.db.module.net",
								Resources: []*tfjson.StateResource{
									managed("module.db.module.net.azurerm_subnet.sn", "azurerm_subnet", "sn",
										"registry.terraform.io/hashicorp/azurerm",
										map[string]any{"name": "sn-1", "location": "eastus"}),
								},
							},
						},
					},
				},
			}}},
			want: []*models.ServiceComponentHealthResource{
				{
					Provider: providerAWS,
					Kind:     "aws_vpc",
					Name:     "vpc-1",
					Health:   healthUnknown,
					Details:  `{"terraform":{"address":"aws_vpc.main","id":"vpc-1"}}`,
				},
				{
					Provider: providerGCP,
					Kind:     "google_sql_database_instance",
					Name:     "pg-prod",
					Health:   healthUnknown,
					Details:  `{"terraform":{"address":"module.db.google_sql_database_instance.pg","region":"us-central1"}}`,
				},
				{
					Provider: providerAzure,
					Kind:     "azurerm_subnet",
					Name:     "sn-1",
					Health:   healthUnknown,
					Details:  `{"terraform":{"address":"module.db.module.net.azurerm_subnet.sn","region":"eastus"}}`,
				},
			},
		},
		{
			name: "provider_derived_from_type_when_unset",
			state: &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{
				Resources: []*tfjson.StateResource{
					managed("aws_iam_role.r", "aws_iam_role", "r", "", map[string]any{"name": "role-a"}),
					managed("google_storage_bucket.b", "google_storage_bucket", "b", "", map[string]any{"name": "b-1"}),
					managed("azurerm_resource_group.rg", "azurerm_resource_group", "rg", "", map[string]any{"name": "rg-1"}),
					managed("docker_image.i", "docker_image", "i", "", map[string]any{"name": "img"}),
				},
			}}},
			want: []*models.ServiceComponentHealthResource{
				{
					Provider: providerAWS,
					Kind:     "aws_iam_role",
					Name:     "role-a",
					Health:   healthUnknown,
					Details:  `{"terraform":{"address":"aws_iam_role.r"}}`,
				},
				{
					Provider: providerGCP,
					Kind:     "google_storage_bucket",
					Name:     "b-1",
					Health:   healthUnknown,
					Details:  `{"terraform":{"address":"google_storage_bucket.b"}}`,
				},
				{
					Provider: providerAzure,
					Kind:     "azurerm_resource_group",
					Name:     "rg-1",
					Health:   healthUnknown,
					Details:  `{"terraform":{"address":"azurerm_resource_group.rg"}}`,
				},
			},
		},
		{
			name: "indexed_instances_stay_distinct",
			state: &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{
				Resources: []*tfjson.StateResource{
					func() *tfjson.StateResource {
						r := managed("aws_instance.web[0]", "aws_instance", "web",
							"registry.terraform.io/hashicorp/aws", nil)
						r.Index = float64(0)
						return r
					}(),
					func() *tfjson.StateResource {
						r := managed("aws_instance.web[1]", "aws_instance", "web",
							"registry.terraform.io/hashicorp/aws", nil)
						r.Index = float64(1)
						r.Tainted = true
						return r
					}(),
				},
			}}},
			want: []*models.ServiceComponentHealthResource{
				{
					Provider: providerAWS,
					Kind:     "aws_instance",
					Name:     "web[0]",
					Health:   healthUnknown,
					Details:  `{"terraform":{"address":"aws_instance.web[0]"}}`,
				},
				{
					Provider: providerAWS,
					Kind:     "aws_instance",
					Name:     "web[1]",
					Health:   healthUnknown,
					Details:  `{"terraform":{"address":"aws_instance.web[1]","tainted":true}}`,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := terraformResourceRows(tc.state)
			if len(tc.want) == 0 {
				assert.Empty(t, got)
				return
			}
			require.Len(t, got, len(tc.want))
			for i := range tc.want {
				assert.Equal(t, *tc.want[i], *got[i])
				assert.Empty(t, got[i].Namespace)
				assert.Empty(t, got[i].Message)
			}
		})
	}
}

func TestTerraformProvider(t *testing.T) {
	newProvider := func() *TerraformProvider {
		return NewTerraformProvider(TerraformProviderParams{L: zap.NewNop()})
	}

	t.Run("records_and_replaces_per_component", func(t *testing.T) {
		p := newProvider()
		assert.Empty(t, p.Resources("cmp-1"))
		assert.Empty(t, p.ComponentIDs())

		p.Set("cmp-1", &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{
			Resources: []*tfjson.StateResource{
				managed("aws_vpc.main", "aws_vpc", "main", "registry.terraform.io/hashicorp/aws",
					map[string]any{"id": "vpc-1"}),
			},
		}}})
		require.Len(t, p.Resources("cmp-1"), 1)
		assert.Equal(t, []string{"cmp-1"}, p.ComponentIDs())

		// a destroy apply leaves an empty state and clears the rows
		p.Set("cmp-1", &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{}}})
		assert.Empty(t, p.Resources("cmp-1"))
	})

	t.Run("ignores_empty_component_id", func(t *testing.T) {
		p := newProvider()
		p.Set("", &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{
			Resources: []*tfjson.StateResource{
				managed("aws_vpc.main", "aws_vpc", "main", "registry.terraform.io/hashicorp/aws", nil),
			},
		}}})
		assert.Empty(t, p.ComponentIDs())
	})

	t.Run("bounds_to_server_cap", func(t *testing.T) {
		resources := make([]*tfjson.StateResource, 0, maxResourcesPerComponent+10)
		for i := 0; i < maxResourcesPerComponent+10; i++ {
			resources = append(resources, managed("aws_vpc.main", "aws_vpc", "main",
				"registry.terraform.io/hashicorp/aws", map[string]any{"id": string(rune('a' + i%26))}))
		}

		p := newProvider()
		p.Set("cmp-1", &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{Resources: resources}}})
		assert.Len(t, p.Resources("cmp-1"), maxResourcesPerComponent)
	})
}
