package build

import (
	"github.com/powertoolsdev/go-workflows-meta/prefix"
)

const (
	phaseName string = "build"
)

func (p *planner) Prefix() string {
	return prefix.DeploymentPhasePath(p.Metadata.OrgShortId,
		p.Metadata.AppShortId,
		p.Component.Id,
		p.Metadata.DeploymentShortId,
		phaseName,
	)
}
