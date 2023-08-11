package worker

import "github.com/powertoolsdev/mono/services/ctl-api/internal"

type Workflows struct {
	cfg *internal.Config
}

func NewWorkflows(cfg *internal.Config) *Workflows {
	return &Workflows{}
}
