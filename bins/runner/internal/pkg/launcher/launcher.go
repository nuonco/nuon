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

// PrepareSpec describes the image a job is about to run steps in. Prepare is
// called once per job so the pull is shared by every step.
type PrepareSpec struct {
	Image        string
	PullUsername string
	PullPassword string

	// LeaseID scopes the lease taken on the image so host garbage collection
	// cannot remove it mid-job. Use the job execution ID.
	LeaseID string

	// PullLog receives image-pull progress (docker writes it to stderr, but it
	// is not an error and not the action's output — keep it separate so it logs
	// at INFO and isn't tagged as command output).
	PullLog io.Writer
}

// RunSpec fully describes an image-backed action container invocation. Env is
// the ONLY environment the container receives — the host/runner environment is
// never inherited (an RFC security requirement).
type RunSpec struct {
	Image         string
	ContainerName string

	Mounts []Mount

	// Command is the container entrypoint + args (the mounted supervisor).
	Command []string

	Env    map[string]string
	Labels map[string]string

	Memory string

	// CPUShares is a relative CPU weight (docker's default is 1024), used
	// instead of a hard quota so an action can consume idle CPU but yields to
	// the management and install processes under contention.
	CPUShares int
	PidsLimit int

	Stdout io.Writer
	Stderr io.Writer
}

// Launcher pulls and runs an image-backed action container.
type Launcher interface {
	// Prepare makes the image available on the host and leases it for the job,
	// so every step reuses one pull and garbage collection leaves it alone
	// until Release.
	Prepare(ctx context.Context, spec PrepareSpec) error

	// Release drops the lease Prepare took. The image stays in the host cache
	// for later runs; only garbage collection removes images.
	Release(leaseID string)

	Run(ctx context.Context, spec RunSpec) error
}
