// Package launcher runs an image-backed action's container. Only a docker
// backend exists today (used by the mng process on VM runners), but the
// Launcher interface is the seam a future Kubernetes Pod/Job backend would
// implement.
package launcher

import (
	"context"
	"io"
)

// Mount is a host-path to container-path bind mount.
type Mount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

// RunSpec fully describes an image-backed action container invocation. Env is
// the ONLY environment the container receives — the host/runner environment is
// never inherited (an RFC security requirement).
type RunSpec struct {
	Image         string
	PullUsername  string
	PullPassword  string
	ContainerName string

	Mounts []Mount

	// Command is the container entrypoint + args (the mounted supervisor).
	Command []string

	Env    map[string]string
	Labels map[string]string

	Memory    string
	CPUs      string
	PidsLimit int

	Stdout io.Writer
	Stderr io.Writer
}

// Launcher pulls and runs an image-backed action container.
type Launcher interface {
	Run(ctx context.Context, spec RunSpec) error
}
