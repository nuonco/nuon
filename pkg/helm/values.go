package helm

import (
	"fmt"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/strvals"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
)

func ChartValues(values []string, helmSet []plantypes.HelmValue) (map[string]interface{}, error) {
	// Next get all our set configs
	base := map[string]interface{}{}

	// First merge all our values from YAML documents.
	for _, values := range values {
		if values == "" {
			continue
		}

		currentVals, err := chartutil.ReadValues([]byte(values))
		if err != nil {
			return nil, errors.Wrap(err, "unable to read values")
		}

		base = chartutil.CoalesceTables(base, currentVals.AsMap())
	}

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
