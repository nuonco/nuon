package iam

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/workers-orgs/internal"
)

const (
	defaultActivityTimeout time.Duration = time.Second * 10
)

func defaultIAMPath(orgID string) string {
	return fmt.Sprintf("/orgs/%s/", orgID)
}

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg internal.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg internal.Config
}
