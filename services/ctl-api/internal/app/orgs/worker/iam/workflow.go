package orgiam

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

const (
	defaultActivityTimeout time.Duration = time.Hour * 1
)

func defaultIAMPath(orgID string) string {
	return fmt.Sprintf("/orgs/%s/", orgID)
}

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg internal.Config) Wkflow {
	return Wkflow{
		cfg: cfg,
	}
}

type Wkflow struct {
	cfg internal.Config
}
