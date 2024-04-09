package domains

import "github.com/powertoolsdev/mono/pkg/shortid"

func NewAppID() string {
	return shortid.NewNanoID("app")
}

func NewAppSecretID() string {
	return shortid.NewNanoID("aps")
}

func NewArtifactID() string {
	return shortid.NewNanoID("art")
}

func NewAWSAccountID() string {
	return shortid.NewNanoID("aws")
}

func NewAzureAccountID() string {
	return shortid.NewNanoID("azu")
}

func NewBuildID() string {
	return shortid.NewNanoID("bld")
}

func NewCanaryID() string {
	return shortid.NewNanoID("can")
}

func NewConfigID() string {
	return shortid.NewNanoID("cfg")
}

func NewComponentID() string {
	return shortid.NewNanoID("cmp")
}

func NewDeploymentID() string {
	return shortid.NewNanoID("dpl")
}

func NewDeployID() string {
	return shortid.NewNanoID("dpl")
}

func NewDomainID() string {
	return shortid.NewNanoID("dom")
}

func NewEventID() string {
	return shortid.NewNanoID("eve")
}

func NewRunID() string {
	return shortid.NewNanoID("run")
}

func NewInstallID() string {
	return shortid.NewNanoID("inl")
}

func NewInstanceID() string {
	return shortid.NewNanoID("ins")
}

func NewMigrationID() string {
	return shortid.NewNanoID("mig")
}

func NewHealthCheck() string {
	return shortid.NewNanoID("hlt")
}

func NewOrgID() string {
	return shortid.NewNanoID("org")
}

func NewVCSConnectionID() string {
	return shortid.NewNanoID("vcs")
}

func NewVCSID() string {
	return shortid.NewNanoID("vcs")
}

func NewSandboxID() string {
	return shortid.NewNanoID("snb")
}

func NewSandboxReleaseID() string {
	return shortid.NewNanoID("snr")
}

func NewSecretID() string {
	return shortid.NewNanoID("sec")
}

func NewUserTokenID() string {
	return shortid.NewNanoID("tok")
}

func NewIntegrationUserID() string {
	return shortid.NewNanoID("int")
}

func NewReleaseID() string {
	return shortid.NewNanoID("rel")
}

func NewUserID() string {
	return shortid.NewNanoID("usr")
}
