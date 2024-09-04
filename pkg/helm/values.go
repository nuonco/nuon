package helm

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/strvals"
)

func ChartValues(values []string, helmSet []configs.HelmSet) (map[string]interface{}, error) {
	base := map[string]interface{}{}

	// First merge all our values from YAML documents.
	for _, values := range values {
		if values == "" {
			continue
		}

		currentMap := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(values), &currentMap); err != nil {
			return nil, fmt.Errorf("---> %v %s", err, values)
		}

		base = generics.MergeMaps(base, currentMap)
	}

	// Next get all our set configs
	for _, set := range helmSet {
		name := set.Name
		value := set.Value
		valueType := set.Type

		switch valueType {
		case "auto", "":
			if err := strvals.ParseInto(fmt.Sprintf("%s=%s", name, value), base); err != nil {
				return nil, fmt.Errorf("failed parsing key %q with value %s, %s", name, value, err)
			}
		case "string":
			if err := strvals.ParseIntoString(fmt.Sprintf("%s=%s", name, value), base); err != nil {
				return nil, fmt.Errorf("failed parsing key %q with value %s, %s", name, value, err)
			}
		default:
			return nil, fmt.Errorf("unexpected type: %s", valueType)
		}
	}

	return base, nil
}
