package configs

import (
	waypointv1 "github.com/hashicorp/waypoint/pkg/server/gen"
)

// This package maps between nuon components and waypoint configs (for builds)
type Builder interface {
	Render() ([]byte, waypointv1.Hcl_Format, error)
}
