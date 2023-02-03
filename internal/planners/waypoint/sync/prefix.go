package sync

import "github.com/powertoolsdev/go-workflows-meta/prefix"

const (
	// NOTE(jm): this phase is used to control the s3 path we emit information to. Future phases will be include
	// `sync-helm-chart` and `sync-terraform-module`
	phaseName string = "sync-container-image"
)

func (p *planner) getPrefix() string {
	return prefix.InstancePhasePath(p.Metadata.OrgShortId,
		p.Metadata.AppShortId,
		p.Component.Name,
		p.Metadata.DeploymentShortId,
		p.Metadata.InstallShortId,
		phaseName,
	)
}
