package app

import "time"

const (
	MinBuildTimeout = time.Second * 1
	MaxBuildTimeout = time.Hour * 1

	MinDeployTimeout = time.Second * 1
	MaxDeployTimeout = time.Hour * 1

	MaxAutoRetries = 20

	MinHealthStabilizationWindow     = time.Second * 1
	MaxHealthStabilizationWindow     = time.Hour * 1
	DefaultHealthStabilizationWindow = time.Minute * 3
)
