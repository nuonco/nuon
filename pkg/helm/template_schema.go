package helm

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mitchellh/copystructure"
	"github.com/santhosh-tekuri/jsonschema/v6"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
)

// Helm's schema compiler can read host files and fetch URLs. Before entering it,
// compile the same active schemas with no external resource loader. Successful
// compilation proves their references (including metaschemas and rebased IDs)
// are self-contained; Helm still performs its normal values validation afterward.
func ValidateSchemaResources(ch *chart.Chart, values map[string]interface{}) error {
	active, err := copyPolicyChart(ch)
	if err != nil {
		return err
	}
	// ProcessDependencies mutates aliases, enabled dependencies and import-values.
	// Use a copy so the real install processes the original exactly once.
	if err := chartutil.ProcessDependencies(active, values); err != nil {
		return err
	}
	return checkActiveSchemaResources(active)
}

func copyPolicyChart(ch *chart.Chart) (*chart.Chart, error) {
	cloned := *ch
	metadata, err := copystructure.Copy(ch.Metadata)
	if err != nil {
		return nil, err
	}
	cloned.Metadata = metadata.(*chart.Metadata)
	values, err := copystructure.Copy(ch.Values)
	if err != nil {
		return nil, err
	}
	cloned.Values = values.(map[string]interface{})
	cloned.SetDependencies()
	for _, child := range ch.Dependencies() {
		copied, err := copyPolicyChart(child)
		if err != nil {
			return nil, err
		}
		cloned.AddDependency(copied)
	}
	return &cloned, nil
}

func checkActiveSchemaResources(ch *chart.Chart) error {
	if len(ch.Schema) > 0 {
		if err := checkSingleSchemaResources(ch.Schema); err != nil {
			return fmt.Errorf("chart %s: %w", ch.Name(), err)
		}
	}
	for _, child := range ch.Dependencies() {
		if err := checkActiveSchemaResources(child); err != nil {
			return err
		}
	}
	return nil
}

func checkSingleSchemaResources(data []byte) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("unable to compile Helm schema: %v", r)
		}
	}()
	schema, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return err
	}
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(policySchemaLoader{})
	// Match Helm's base URI, draft selection and compiler defaults exactly.
	const address = "file:///values.schema.json"
	if err := compiler.AddResource(address, schema); err != nil {
		return err
	}
	_, err = compiler.Compile(address)
	return err
}

type policySchemaLoader struct{}

func (policySchemaLoader) Load(address string) (any, error) {
	if strings.HasPrefix(address, "urn:") {
		return true, nil
	}
	return nil, fmt.Errorf("external schema resource is not allowed for Helm rendering: %s", address)
}
