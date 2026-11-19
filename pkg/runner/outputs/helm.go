package outputs

import (
	"github.com/databus23/helm-diff/v3/manifest"
)

// HelmOutputs is a helper method for fetching a json helm manifest representation to write as an output.
func HelmOutputs(manifestStr, ns string) (map[string]interface{}, error) {
	mapping := manifest.Parse(manifestStr, ns, true)

	return map[string]interface{}{
		"manifest":  manifestStr,
		"resources": mapping,
	}, nil
}
