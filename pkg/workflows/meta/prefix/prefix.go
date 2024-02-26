package prefix

import (
	"fmt"
	"path/filepath"
)

type appConfig struct {
	OrgID string
	AppID string
}

func (i appConfig) toPath() string {
	base := fmt.Sprintf("org=%s/app=%s",
		i.OrgID,
		i.AppID)
	return base
}

// InstanceStatePath returns the prefix for an instance's state - meaning long lived files that need to persist through
// runs, such as a terraform state file.
func AppConfigPath(orgID, appID string) string {
	return appConfig{
		OrgID: orgID,
		AppID: appID,
	}.toPath()
}

type instanceState struct {
	OrgID       string
	AppID       string
	ComponentID string
	InstallID   string
}

func (i instanceState) toPath() string {
	base := fmt.Sprintf("org=%s/app=%s/component=%s/install=%s",
		i.OrgID,
		i.AppID,
		i.ComponentID,
		i.InstallID)
	return base
}

// InstanceStatePath returns the prefix for an instance's state - meaning long lived files that need to persist through
// runs, such as a terraform state file.
func InstanceStatePath(orgID, appID, componentID, installID string) string {
	return instanceState{
		OrgID:       orgID,
		AppID:       appID,
		ComponentID: componentID,
		InstallID:   installID,
	}.toPath()
}

// InstanceOutputPath returns the prefix for an instance's output - meaning long lived files that need to persist through
// runs, such as component output values.
var InstanceOutputPath = InstanceStatePath

type instance struct {
	OrgID       string
	AppID       string
	ComponentID string
	DeployID    string
	InstallID   string
	Phase       string
}

func (i instance) toPath() string {
	base := fmt.Sprintf("org=%s/app=%s/component=%s/install=%s/deploy=%s",
		i.OrgID,
		i.AppID,
		i.ComponentID,
		i.InstallID,
		i.DeployID,
	)
	if i.Phase != "" {
		base = filepath.Join(base, fmt.Sprintf("phase=%s", i.Phase))
	}

	return base
}

// InstancePath returns the prefix for an instance
func InstancePath(orgID, appID, componentID, deployID, installID string) string {
	return instance{
		OrgID:       orgID,
		AppID:       appID,
		ComponentID: componentID,
		DeployID:    deployID,
		InstallID:   installID,
	}.toPath()
}

// InstancePhasePath returns the prefix for an instance's phase
func InstancePhasePath(orgID, appID, componentName, deployID, installID, phase string) string {
	return instance{
		OrgID:       orgID,
		AppID:       appID,
		ComponentID: componentName,
		DeployID:    deployID,
		InstallID:   installID,
		Phase:       phase,
	}.toPath()
}

type componentInstall struct {
	OrgID       string
	AppID       string
	ComponentID string
	InstallID   string
}

func (ci componentInstall) toPath() string {
	return fmt.Sprintf("org=%s/app=%s/component=%s/install=%s", ci.OrgID, ci.AppID, ci.ComponentID, ci.InstallID)
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

type build struct {
	OrgID       string
	AppID       string
	ComponentID string
	BuildID     string
}

func (b build) toPath() string {
	base := fmt.Sprintf("org=%s/app=%s/component=%s/build=%s",
		b.OrgID,
		b.AppID,
		b.ComponentID,
		b.BuildID)

	return base
}

// BuildPath returns the prefix for a build
func BuildPath(orgID, appID, componentID, buildID string) string {
	return build{
		OrgID:       orgID,
		AppID:       appID,
		ComponentID: componentID,
		BuildID:     buildID,
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

type orgComponent struct {
	OrgID         string
	ComponentName string
}

// OrgPath returns the prefix for an org
func OrgComponentPath(orgID string, componentName string) string {
	return orgComponent{
		OrgID:         orgID,
		ComponentName: componentName,
	}.toPath()
}

func (o orgComponent) toPath() string {
	return fmt.Sprintf("org=%s/component=%s", o.OrgID, o.ComponentName)
}

type installationStatic struct {
	OrgID     string
	AppID     string
	InstallID string
}

func InstallationStaticPath(orgID, appID, installID string) string {
	return installationStatic{
		OrgID:     orgID,
		AppID:     appID,
		InstallID: installID,
	}.toPath()
}

func (s installationStatic) toPath() string {
	return fmt.Sprintf("org=%s/app=%s/install=%s",
		s.OrgID,
		s.AppID,
		s.InstallID,
	)
}

type installation struct {
	OrgID     string
	AppID     string
	InstallID string
	RunID     string
}

func InstallationPath(orgID, appID, installID, runID string) string {
	return installation{
		OrgID:     orgID,
		AppID:     appID,
		InstallID: installID,
		RunID:     runID,
	}.toPath()
}

func (s installation) toPath() string {
	return fmt.Sprintf("org=%s/app=%s/install=%s/run=%s",
		s.OrgID,
		s.AppID,
		s.InstallID,
		s.RunID,
	)
}

func SecretsPath(orgID, appID, componentID, installID string) string {
	return componentInstall{
		OrgID:       orgID,
		AppID:       appID,
		ComponentID: componentID,
		InstallID:   installID,
	}.toPath()
}
