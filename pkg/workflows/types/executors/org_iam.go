package executors

const (
	ProvisionIAMWorkflowName   string = "ProvisionIAM"
	DeprovisionIAMWorkflowName string = "DeprovisionIAM"
)

type ProvisionIAMRequest struct {
	OrgId       string
	Reprovision bool
}

type ProvisionIAMResponse struct {
	DeploymentsRoleArn   string
	InstallationsRoleArn string
	OdrRoleArn           string
	InstancesRoleArn     string
	InstallerRoleArn     string
	OrgsRoleArn          string
	SecretsRoleArn       string
}

type DeprovisionIAMRequest struct {
	OrgId string
}

type DeprovisionIAMResponse struct{}
