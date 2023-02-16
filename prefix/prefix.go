package prefix

import (
	"fmt"
	"path/filepath"
)

type instance struct {
	OrgID         string
	AppID         string
	ComponentName string
	DeploymentID  string
	InstallID     string
	Phase         string
}

func (i instance) toPath() string {
	base := fmt.Sprintf("org=%s/app=%s/component=%s/deployment=%s/install=%s",
		i.OrgID,
		i.AppID,
		i.ComponentName,
		i.DeploymentID,
		i.InstallID)
	if i.Phase != "" {
		base = filepath.Join(base, fmt.Sprintf("phase=%s", i.Phase))
	}

	return base
}

// InstancePath returns the prefix for an instance
func InstancePath(orgID, appID, componentName, deploymentID, installID string) string {
	return instance{
		OrgID:         orgID,
		AppID:         appID,
		ComponentName: componentName,
		DeploymentID:  deploymentID,
		InstallID:     installID,
	}.toPath()
}

// InstancePhasePath returns the prefix for an instance's phase
func InstancePhasePath(orgID, appID, componentName, deploymentID, installID, phase string) string {
	return instance{
		OrgID:         orgID,
		AppID:         appID,
		ComponentName: componentName,
		DeploymentID:  deploymentID,
		InstallID:     installID,
		Phase:         phase,
	}.toPath()
}

type install struct {
	OrgID     string
	AppID     string
	InstallID string
}

func (i install) toPath() string {
	return fmt.Sprintf("org=%s/app=%s/install=%s",
		i.OrgID,
		i.AppID,
		i.InstallID)
}

// InstallPath returns the prefix for an instance
func InstallPath(orgID, appID, installID string) string {
	return install{
		OrgID:     orgID,
		AppID:     appID,
		InstallID: installID,
	}.toPath()
}

type deployment struct {
	OrgID         string
	AppID         string
	ComponentName string
	DeploymentID  string
	Phase         string
}

func (d deployment) toPath() string {
	base := fmt.Sprintf("org=%s/app=%s/component=%s/deployment=%s",
		d.OrgID,
		d.AppID,
		d.ComponentName,
		d.DeploymentID)
	if d.Phase != "" {
		base = filepath.Join(base, fmt.Sprintf("phase=%s", d.Phase))
	}

	return base
}

// DeploymentPhasePath returns the prefix for a deployment
func DeploymentPhasePath(orgID, appID, componentName, deploymentID, phase string) string {
	return deployment{
		OrgID:         orgID,
		AppID:         appID,
		ComponentName: componentName,
		DeploymentID:  deploymentID,
		Phase:         phase,
	}.toPath()
}

// DeploymentPath returns the prefix for a deployment
func DeploymentPath(orgID, appID, componentName, deploymentID string) string {
	return deployment{
		OrgID:         orgID,
		AppID:         appID,
		ComponentName: componentName,
		DeploymentID:  deploymentID,
	}.toPath()
}

type app struct {
	OrgID string
	AppID string
}

func (a app) toPath() string {
	return fmt.Sprintf("org=%s/app=%s",
		a.OrgID, a.AppID)
}

// AppPath returns the prefix for an org
func AppPath(orgID, appID string) string {
	return app{
		OrgID: orgID,
		AppID: appID,
	}.toPath()
}

type org struct {
	OrgID string
}

// OrgPath returns the prefix for an org
func OrgPath(orgID string) string {
	return org{
		OrgID: orgID,
	}.toPath()
}

func (o org) toPath() string {
	return fmt.Sprintf("org=%s", o.OrgID)
}

type sandbox struct {
	OrgID          string
	AppID          string
	InstallID      string
	SandboxName    string
	SandboxVersion string
}

func SandboxPath(orgID, appID, installID, sandboxName, sandboxVersion string) string {
	return sandbox{
		OrgID:          orgID,
		AppID:          appID,
		InstallID:      installID,
		SandboxName:    sandboxName,
		SandboxVersion: sandboxVersion,
	}.toPath()
}

func (s sandbox) toPath() string {
	return fmt.Sprintf("org=%s/app=%s/install=%s/sandbox=%s/version=%s",
		s.OrgID,
		s.AppID,
		s.InstallID,
		s.SandboxName,
		s.SandboxVersion,
	)
}
