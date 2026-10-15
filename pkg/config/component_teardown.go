package config

import (
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/terraform/teardown"
)

// ComponentTeardownConfig configures how a component's resources are handled
// when the component is torn down.
type ComponentTeardownConfig struct {
	Retain []string `mapstructure:"retain,omitempty" toml:"retain,omitempty" features:"template" nuonhash:"omitempty"`
}

func (c ComponentTeardownConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("retain").Short("resources to leave in place on teardown").
		Long("Terraform resource or module addresses that are removed from state instead of destroyed when the component is torn down. " +
			"An address without an index selects every instance of the resource, and a module address selects every resource in the module. " +
			"Supports Nuon templating; entries that render to an empty string are ignored").
		Example([]string{"aws_rds_cluster.main", "aws_s3_bucket.data[\"logs\"]", "module.buckets"})
}

// Validate checks the syntax of each retain address. Templated entries only
// have their template checked; the address is checked once rendered.
func (c *ComponentTeardownConfig) Validate() error {
	if c == nil {
		return nil
	}

	seen := make(map[string]struct{}, len(c.Retain))
	for idx, addr := range c.Retain {
		if _, ok := seen[addr]; ok {
			return fmt.Errorf("teardown.retain[%d]: duplicate address %q", idx, addr)
		}
		seen[addr] = struct{}{}

		if strings.Contains(addr, "{{") {
			if err := render.ValidateTextTemplate(addr); err != nil {
				return fmt.Errorf("teardown.retain[%d]: invalid template: %w", idx, err)
			}
			continue
		}
		if err := teardown.ValidateAddress(addr); err != nil {
			return fmt.Errorf("teardown.retain[%d]: %w", idx, err)
		}
	}

	return nil
}
