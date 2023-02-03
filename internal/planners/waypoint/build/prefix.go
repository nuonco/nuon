package build

import "github.com/powertoolsdev/go-workflows-meta/prefix"

const (
	phaseName string = "build"
)

func (p *planner) getPrefix() string {
	return prefix.DeploymentPhasePath(p.Metadata.OrgShortId,
		p.Metadata.AppShortId,
		p.Component.Name,
		p.Metadata.DeploymentShortId,
		phaseName,
	)
}
