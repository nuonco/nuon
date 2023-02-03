package build

import "fmt"

func (p *planner) getPrefix() string {
	return fmt.Sprintf("org=%s/app=%s/component=%s/deployment=%s/",
		p.Metadata.OrgShortId,
		p.Metadata.AppShortId,
		p.Component.Name,
		p.Metadata.DeploymentShortId)
}
