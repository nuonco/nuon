package airgap

import (
	"bytes"
	"fmt"
	"regexp"
)

// deploymentSuffixPattern keeps suffixes safe for every naming context a
// derived name lands in (IAM role names, EKS cluster names, ECR repository
// names, Route53 labels, S3 keys).
var deploymentSuffixPattern = regexp.MustCompile(`^[a-z0-9]{1,8}$`)

// minDeploymentIDPrefix is how much of the original install ID must survive
// the splice so derived names stay recognizable and unlikely to collide with
// a different install's truncation.
const minDeploymentIDPrefix = 12

// DeploymentInstallID derives a deployment-scoped install ID by replacing the
// tail of the frozen install ID with "-<suffix>". The result is exactly the
// same length as the original: every physical name in the exported plans and
// stack templates is derived by concatenating the install ID, and several of
// them (the EKS module's IAM role name_prefix in particular) already sit at
// their length limit, so the substitution must never grow a name.
func DeploymentInstallID(installID, suffix string) (string, error) {
	if !deploymentSuffixPattern.MatchString(suffix) {
		return "", fmt.Errorf("deployment id %q must be 1-8 lowercase letters or digits", suffix)
	}
	keep := len(installID) - len(suffix) - 1
	if keep < minDeploymentIDPrefix {
		return "", fmt.Errorf("install ID %q is too short to carry deployment id %q", installID, suffix)
	}
	return installID[:keep] + "-" + suffix, nil
}

// ApplyDeploymentID rewrites every occurrence of the envelope's frozen
// install ID to the deployment-scoped ID, in each step's composite plan and
// the app config, then adopts the new ID as the envelope's install ID. The
// rewrite is a raw token substitution: install IDs are plain alphanumeric
// strings, so replacing them inside JSON never breaks encoding, and because
// the substitution is applied uniformly the plans stay internally consistent
// (reference snapshots, rendered vars, and role ARNs all move together, so
// the runner's late-binding value matches keep working).
func (e *Envelope) ApplyDeploymentID(suffix string) (string, error) {
	deploymentID, err := DeploymentInstallID(e.InstallID, suffix)
	if err != nil {
		return "", err
	}
	old, fresh := []byte(e.InstallID), []byte(deploymentID)
	for i := range e.Steps {
		e.Steps[i].CompositePlan = bytes.ReplaceAll(e.Steps[i].CompositePlan, old, fresh)
	}
	for i := range e.Actions {
		e.Actions[i].CompositePlan = bytes.ReplaceAll(e.Actions[i].CompositePlan, old, fresh)
	}
	for i := range e.Drift {
		e.Drift[i].CompositePlan = bytes.ReplaceAll(e.Drift[i].CompositePlan, old, fresh)
	}
	if len(e.AppConfig) > 0 {
		e.AppConfig = bytes.ReplaceAll(e.AppConfig, old, fresh)
	}
	e.InstallID = deploymentID
	return deploymentID, nil
}
