package domains

import "github.com/powertoolsdev/mono/pkg/common/shortid"

func NewAppID() string {
	return shortid.NewNanoID("app")
}

func NewArtifactID() string {
	return shortid.NewNanoID("art")
}

func NewAWSSettingsID() string {
	return shortid.NewNanoID("aws")
}

func NewBuildID() string {
	return shortid.NewNanoID("bld")
}

func NewCanaryID() string {
	return shortid.NewNanoID("can")
}

func NewComponentID() string {
	return shortid.NewNanoID("cmp")
}

func NewDeploymentID() string {
	return shortid.NewNanoID("dpl")
}

func NewDomainID() string {
	return shortid.NewNanoID("dom")
}

func NewInstallID() string {
	return shortid.NewNanoID("inl")
}

func NewInstanceID() string {
	return shortid.NewNanoID("ins")
}

func NewOrgID() string {
	return shortid.NewNanoID("org")
}

func NewSandboxID() string {
	return shortid.NewNanoID("snb")
}

func NewSecretID() string {
	return shortid.NewNanoID("sec")
}

func NewUserID() string {
	return shortid.NewNanoID("usr")
}
