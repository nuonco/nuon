package gcp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"text/template"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// GCPTemplateInput extends TemplateInput with pre-marshaled GCP IAM permission lists.
type GCPTemplateInput struct {
	*stacks.TemplateInput
	ProvisionPermissions   string
	MaintenancePermissions string
	DeprovisionPermissions string
	BreakGlassPermissions  string
	HasBreakGlass          bool

	ProvisionPredefinedRole   string
	MaintenancePredefinedRole string
	DeprovisionPredefinedRole string
	BreakGlassPredefinedRole  string
}

func Render(inputs *stacks.TemplateInput) ([]byte, string, error) {
	t, err := template.New("gcp-stack").Parse(tmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to parse gcp template")
	}

	prov, maint, deprov, bg, provPredefined, maintPredefined, deprovPredefined, bgPredefined := extractGCPPermissions(inputs.AppCfg)
	gcpInputs := &GCPTemplateInput{
		TemplateInput:             inputs,
		ProvisionPermissions:      prov,
		MaintenancePermissions:    maint,
		DeprovisionPermissions:    deprov,
		BreakGlassPermissions:     bg,
		HasBreakGlass:             bg != "[]",
		ProvisionPredefinedRole:   provPredefined,
		MaintenancePredefinedRole: maintPredefined,
		DeprovisionPredefinedRole: deprovPredefined,
		BreakGlassPredefinedRole:  bgPredefined,
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, gcpInputs)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to execute gcp template")
	}

	// Wrap tfvars in a JSON envelope so it can be stored in the jsonb column.
	// The raw tfvars text is HCL, not valid JSON.
	envelope := map[string]string{"tfvars": buf.String()}
	res, err := json.Marshal(envelope)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal gcp tfvars envelope")
	}

	hash := sha256.Sum256(res)
	checksum := hex.EncodeToString(hash[:])

	return res, checksum, nil
}

// extractGCPPermissions reads GCP IAM permissions from the app config for each role type.
// Returns empty arrays for any role type that has no GCP permissions configured.
func extractGCPPermissions(appCfg *app.AppConfig) (provision, maintenance, deprovision, breakGlass, provPredefined, maintPredefined, deprovPredefined, bgPredefined string) {
	provision = "[]"
	maintenance = "[]"
	deprovision = "[]"
	breakGlass = "[]"
	provPredefined = ""
	maintPredefined = ""
	deprovPredefined = ""
	bgPredefined = ""

	if appCfg == nil {
		return
	}

	allRoles := append(appCfg.PermissionsConfig.Roles, appCfg.BreakGlassConfig.Roles...)
	for _, role := range allRoles {
		if role.CloudPlatform != "gcp" {
			continue
		}

		var perms []string
		var predefinedRole string
		for _, policy := range role.Policies {
			perms = append(perms, policy.GCPPermissions...)
			if policy.GCPPredefinedRole != "" {
				predefinedRole = policy.GCPPredefinedRole
			}
		}
		if len(perms) == 0 && predefinedRole == "" {
			continue
		}

		if len(perms) > 0 {
			b, err := json.Marshal(perms)
			if err != nil {
				continue
			}

			switch role.Type {
			case app.AWSIAMRoleTypeRunnerProvision:
				provision = string(b)
			case app.AWSIAMRoleTypeRunnerMaintenance:
				maintenance = string(b)
			case app.AWSIAMRoleTypeRunnerDeprovision:
				deprovision = string(b)
			case app.AWSIAMRoleTypeBreakGlass, app.AWSIAMRoleTypeRunnerBreakGlass:
				breakGlass = string(b)
			}
		}

		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provPredefined = predefinedRole
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintPredefined = predefinedRole
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovPredefined = predefinedRole
		case app.AWSIAMRoleTypeBreakGlass, app.AWSIAMRoleTypeRunnerBreakGlass:
			bgPredefined = predefinedRole
		}
	}

	return
}
