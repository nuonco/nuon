package plan

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
)

const (
	configFilename string = "waypoint.hcl"
)

func ParseConfig[T any](plan *planv1.Plan) (*T, error) {
	wpPlan := plan.GetWaypointPlan()
	if wpPlan == nil {
		return nil, fmt.Errorf("invalid config, waypoint plan is nil")
	}

	cfgStr := wpPlan.WaypointJob.HclConfig
	fmt.Println(cfgStr)
	var cfg T

	if err := hclsimple.Decode(configFilename, []byte(cfgStr), nil, &cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config from plan: %w", err)
	}

	return &cfg, nil
}
