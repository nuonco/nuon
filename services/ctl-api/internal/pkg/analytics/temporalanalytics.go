package analytics

import (
	"github.com/pkg/errors"

	temporalanalytics "github.com/powertoolsdev/mono/pkg/analytics/temporal"
)

func NewTemporal(params Params) (temporalanalytics.Writer, error) {
	w, err := temporalanalytics.New(params.V,
		temporalanalytics.WithDisable(params.Cfg.DisableAnalytics),
		temporalanalytics.WithSegmentKey(params.Cfg.SegmentWriteKey),
		temporalanalytics.WithLogger(params.L),
		temporalanalytics.WithUserIDFn(temporalUserIDFn),
		temporalanalytics.WithProperties(map[string]interface{}{
			"platform": "ctl-api",
			"env":      params.Cfg.Env,
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get analytics writer")
	}

	return w, nil
}
