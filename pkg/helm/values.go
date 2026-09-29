package helm

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	chartcommon "helm.sh/helm/v4/pkg/chart/common"
	chartutil "helm.sh/helm/v4/pkg/chart/common/util"
	"helm.sh/helm/v4/pkg/strvals"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func ChartValues(values []string, helmSet []plantypes.HelmValue, override string) (map[string]interface{}, error) {
	base := map[string]interface{}{}

	for _, values := range values {
		if values == "" {
			continue
		}

		currentVals, err := chartcommon.ReadValues([]byte(values))
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

	if strings.TrimSpace(override) != "" {
		overrideVals, err := chartcommon.ReadValues([]byte(override))
		if err != nil {
			return nil, errors.Wrap(err, "unable to read install override values")
		}
		base = chartutil.CoalesceTables(overrideVals.AsMap(), base)
	}

	return base, nil
}
