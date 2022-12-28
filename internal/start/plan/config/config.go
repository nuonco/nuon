package config

import (
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
	componentv1 "github.com/powertoolsdev/protos/components/generated/types/component/v1"
	planv1 "github.com/powertoolsdev/protos/deployments/generated/types/plan/v1"
)

// This package maps between nuon components and waypoint configs (for builds)
type Builder interface {
	WithMetadata(*planv1.Metadata)
	WithECRRef(*planv1.ECRRepositoryRef)
	WithComponent(*componentv1.Component)

	Render() ([]byte, waypointv1.Hcl_Format, error)
}
