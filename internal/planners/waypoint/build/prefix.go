package build

import "github.com/powertoolsdev/go-workflows-meta/prefix"

func (p *planner) getPrefix() string {
	return prefix.DeploymentPath(p.Metadata.OrgShortId,
		p.Metadata.AppShortId,
		p.Component.Name,
		p.Metadata.DeploymentShortId)
}
