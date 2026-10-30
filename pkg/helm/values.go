package helm

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/v4/pkg/action"
	chartcommon "helm.sh/helm/v4/pkg/chart/common"
	chartutil "helm.sh/helm/v4/pkg/chart/common/util"
	rcommon "helm.sh/helm/v4/pkg/release/common"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage/driver"
	"helm.sh/helm/v4/pkg/strvals"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
)

func ChartValues(values []string, helmSet []plantypes.HelmValue, override string) (map[string]interface{}, error) {
	// Next get all our set configs
	base := map[string]interface{}{}

	// First merge all our values from YAML documents.
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

	// Finally, apply the install-level values override as the highest-precedence
	// layer. It is coalesced override-authoritative, so it deep-merges over the
	// values files AND the inline --set values and wins on any overlapping key.
	// An empty override is an exact no-op.
	if strings.TrimSpace(override) != "" {
		overrideVals, err := chartcommon.ReadValues([]byte(override))
		if err != nil {
			return nil, errors.Wrap(err, "unable to read install override values")
		}
		base = chartutil.CoalesceTables(overrideVals.AsMap(), base)
	}

	return base, nil
}

// UpgradeValues resolves Helm's default empty-input fallback before schema
// confinement. Run the upgrade with ResetValues so it uses this same selection.
func UpgradeValues(cfg *action.Configuration, last *release.Release, values map[string]interface{}) (map[string]interface{}, error) {
	if len(values) > 0 {
		return values, nil
	}
	current := last
	if last.Info.Status != rcommon.StatusDeployed {
		deployed, err := cfg.Releases.Deployed(last.Name)
		if err != nil {
			if !errors.Is(err, driver.ErrNoDeployedReleases) || (last.Info.Status != rcommon.StatusFailed && last.Info.Status != rcommon.StatusSuperseded) {
				return nil, err
			}
		} else {
			current, err = AsRelease(deployed)
			if err != nil {
				return nil, err
			}
		}
	}
	return current.Config, nil
}
