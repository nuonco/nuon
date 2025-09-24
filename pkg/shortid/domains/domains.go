package domains

import (
	"strings"

	"github.com/powertoolsdev/mono/pkg/shortid"
)

func NewAppID() string {
	return shortid.NewNanoID("app")
}

func NewAppCfgID() string {
	return shortid.NewNanoID("apc")
}

func NewTemporalPayload() string {
	return shortid.NewNanoID("tpp")
}

func NewAppBranchID() string {
	return shortid.NewNanoID("abr")
}

func NewAccountID() string {
	return shortid.NewNanoID("acc")
}

func NewAccountPolicyID() string {
	return shortid.NewNanoID("acp")
}

func NewRoleID() string {
	return shortid.NewNanoID("rol")
}

func NewInstallerID() string {
	return shortid.NewNanoID("int")
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

func NewInstallStackID() string {
	return shortid.NewNanoID("ist")
}

func NewInstallStateID() string {
	return shortid.NewNanoID("isa")
}

func NewInstallStackVersionID() string {
	return shortid.NewNanoID("isv")
}

func NewInstallStackVersionRunID() string {
	return shortid.NewNanoID("isr")
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

func NewInfraTestID() string {
	return shortid.NewNanoID("its")
}

func NewConfigID() string {
	return shortid.NewNanoID("cfg")
}

func NewInstallComponentID() string {
	return shortid.NewNanoID("inc")
}

func NewInstallSandboxID() string {
	return shortid.NewNanoID("isb")
}

func NewComponentID() string {
	return shortid.NewNanoID("cmp")
}

func NewActionID() string {
	return shortid.NewNanoID("act")
}

func NewActionWorkflowID() string {
	return shortid.NewNanoID("acw")
}

func NewInstallActionWorkflowConfigID() string {
	return shortid.NewNanoID("iaw")
}

func NewActionWorkflowConfigID() string {
	return shortid.NewNanoID("acc")
}

func NewActionWorkflowStepConfigID() string {
	return shortid.NewNanoID("acs")
}

func NewActionWorkflowTriggerConfigID() string {
	return shortid.NewNanoID("act")
}

func NewActionRunID() string {
	return shortid.NewNanoID("acr")
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

func NewInstallActionWorkflowRunID() string {
	return shortid.NewNanoID("iar")
}

func NewRunnerOperationID() string {
	return shortid.NewNanoID("rop")
}

func NewSandboxRunID() string {
	return shortid.NewNanoID("sbr")
}

func NewInstallID() string {
	return shortid.NewNanoID("inl")
}

func NewInstallConfigID() string {
	return shortid.NewNanoID("icg")
}

func NewAppRepoID() string {
	return shortid.NewNanoID("apr")
}

func NewWorkflowID() string {
	return shortid.NewNanoID("inw")
}

func NewWorkflowStepID() string {
	return shortid.NewNanoID("iws")
}

func NewWorkflowStepApprovalID() string {
	return shortid.NewNanoID("waa")
}

func NewWorkflowStepApprovalResponseID() string {
	return shortid.NewNanoID("war")
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

func NewNotificationsID() string {
	return shortid.NewNanoID("ntf")
}

func NewOrgID() string {
	return shortid.NewNanoID("org")
}

func NewOtelTraceID() string {
	return shortid.NewNanoID("trc")
}

func NewOtelLogID() string {
	return shortid.NewNanoID("log")
}

func NewOtelMetricSumID() string {
	return shortid.NewNanoID("msu")
}

func NewOtelMetricGaugeID() string {
	return shortid.NewNanoID("mga")
}

func NewOtelMetricHistogramID() string {
	return shortid.NewNanoID("mhi")
}

func NewOtelMetricExponentialHistogramID() string {
	return shortid.NewNanoID("meh")
}

func NewOtelMetricSummaryID() string {
	return shortid.NewNanoID("msr")
}

func NewRunnerID() string {
	return shortid.NewNanoID("run")
}

func NewLogStreamID() string {
	return shortid.NewNanoID("log")
}

func IsRunnerID(val string) bool {
	if !shortid.IsShortID(val) {
		return false
	}

	if !strings.HasPrefix(val, "run") {
		return false
	}

	return true
}

func NewRunnerHealthCheckID() string {
	return shortid.NewNanoID("hlt")
}

func NewRunnerHeartBeatID() string {
	return shortid.NewNanoID("hrt")
}

func NewRunnerJobID() string {
	return shortid.NewNanoID("job")
}

func NewQueueID() string {
	return shortid.NewNanoID("que")
}

func NewQueueSignalID() string {
	return shortid.NewNanoID("qsi")
}

func NewOCIArtifactID() string {
	return shortid.NewNanoID("oci")
}

func NewTerraformWorkspaceID() string {
	return shortid.NewNanoID("tfw")
}

func NewTerraformWorkspaceLockID() string {
	return shortid.NewNanoID("tfl")
}

func NewTerraformWorkspaceStateID() string {
	return shortid.NewNanoID("tfs")
}

func NewTerraformWorkspaceStateJSONID() string {
	return shortid.NewNanoID("tfj")
}

func NewRunnerGroupID() string {
	return shortid.NewNanoID("rgr")
}

func NewRunnerGroupSettingsID() string {
	return shortid.NewNanoID("rgs")
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

func NewWaitListID() string {
	return shortid.NewNanoID("wtl")
}

func NewHelmChartID() string {
	return shortid.NewNanoID("hmc")
}

func NewEndpointAuditID() string {
	return shortid.NewNanoID("epa")
}
