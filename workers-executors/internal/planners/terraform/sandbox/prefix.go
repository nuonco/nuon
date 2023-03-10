package sandbox

import "github.com/powertoolsdev/go-workflows-meta/prefix"

func (p *planner) Prefix() string {
	return prefix.SandboxPath(
		p.sandbox.OrgId,
		p.sandbox.AppId,
		p.sandbox.InstallId,
		p.sandbox.SandboxSettings.Name,
		p.sandbox.SandboxSettings.Version,
	)
}
