package launcher

import (
	"context"
	"io"
)

type Mount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

type PrepareSpec struct {
	Image        string
	PullUsername string
	PullPassword string

	LeaseID string

	PullLog io.Writer
}

type RunSpec struct {
	Image         string
	ContainerName string

	Mounts []Mount

	Command []string
	Workdir string

	Env    map[string]string
	Labels map[string]string

	Memory string

	CPUShares int
	PidsLimit int

	Stdout io.Writer
	Stderr io.Writer
}

type Launcher interface {
	Prepare(ctx context.Context, spec PrepareSpec) error

	Release(leaseID string)

	Run(ctx context.Context, spec RunSpec) error
}
