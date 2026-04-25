package app

import "time"

const (
	// Build timeout bounds (applies to build operations)
	MinBuildTimeout = time.Second * 1
	MaxBuildTimeout = time.Hour * 1

	// Deploy timeout bounds (applies to deploy operations)
	MinDeployTimeout = time.Second * 1
	MaxDeployTimeout = time.Hour * 1

	// Auto retry bounds (applies to deploy auto-retry)
	MaxAutoRetries = 20
)
