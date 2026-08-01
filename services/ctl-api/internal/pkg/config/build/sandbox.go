package build

import (
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
)

type SandboxInput struct {
	AppID       string
	AppConfigID string

	Type             string
	TerraformVersion string
	Runtime          string
	PulumiVersion    string

	PulumiConfig map[string]string
	Variables    map[string]string
	EnvVars      map[string]string

	VariablesFiles []string
	References     []string
	OperationRoles []config.EntityOperationRole

	DriftSchedule                *string
	MaxAutoRetries               *int
	SkipNoops                    *bool
	AutoApproveOnPoliciesPassing *bool

	GithubVCSConfig    *app.ConnectedGithubVCSConfig
	PublicGitVCSConfig *app.PublicGitVCSConfig
}

func SandboxInputFromConfig(sandbox *config.AppSandboxConfig, appID, appConfigID string) SandboxInput {
	variablesFiles := make([]string, 0, len(sandbox.VariablesFiles))
	for _, vf := range sandbox.VariablesFiles {
		variablesFiles = append(variablesFiles, vf.Contents)
	}

	references := make([]string, 0, len(sandbox.References))
	for _, ref := range sandbox.References {
		references = append(references, ref.String())
	}

	return SandboxInput{
		AppID:                        appID,
		AppConfigID:                  appConfigID,
		Type:                         sandbox.Type,
		TerraformVersion:             sandbox.TerraformVersion,
		Runtime:                      sandbox.Runtime,
		PulumiVersion:                sandbox.PulumiVersion,
		PulumiConfig:                 sandbox.PulumiConfig,
		Variables:                    sandbox.VarsMap,
		EnvVars:                      sandbox.EnvVarMap,
		VariablesFiles:               variablesFiles,
		References:                   references,
		OperationRoles:               sandbox.OperationRoles,
		DriftSchedule:                sandbox.DriftSchedule,
		MaxAutoRetries:               sandbox.MaxAutoRetries,
		SkipNoops:                    sandbox.SkipNoops,
		AutoApproveOnPoliciesPassing: sandbox.AutoApproveOnPoliciesPassing,
	}
}

func SandboxConfig(in SandboxInput) (*app.AppSandboxConfig, error) {
	sandboxType, err := ResolveSandboxType(in.Type, in.TerraformVersion, in.Runtime)
	if err != nil {
		return nil, err
	}

	if in.MaxAutoRetries != nil {
		if err := validation.ValidateMaxAutoRetries(*in.MaxAutoRetries); err != nil {
			return nil, err
		}
	}
	if in.DriftSchedule != nil {
		if err := validation.ValidateCronSchedule(*in.DriftSchedule); err != nil {
			return nil, err
		}
	}
	if err := ValidateOperationRoles(in.OperationRoles); err != nil {
		return nil, err
	}

	obj := &app.AppSandboxConfig{
		AppID:                        in.AppID,
		AppConfigID:                  in.AppConfigID,
		PublicGitVCSConfig:           in.PublicGitVCSConfig,
		ConnectedGithubVCSConfig:     in.GithubVCSConfig,
		Type:                         sandboxType,
		TerraformVersion:             in.TerraformVersion,
		Runtime:                      in.Runtime,
		PulumiVersion:                in.PulumiVersion,
		PulumiConfig:                 Hstore(in.PulumiConfig),
		Variables:                    Hstore(in.Variables),
		EnvVars:                      Hstore(in.EnvVars),
		VariablesFiles:               pq.StringArray(sliceOrEmpty(in.VariablesFiles)),
		References:                   pq.StringArray(sliceOrEmpty(in.References)),
		OperationRoles:               OperationRoles(in.OperationRoles),
		MaxAutoRetries:               in.MaxAutoRetries,
		SkipNoops:                    in.SkipNoops,
		AutoApproveOnPoliciesPassing: in.AutoApproveOnPoliciesPassing,
	}

	if in.DriftSchedule != nil {
		obj.DriftSchedule = *in.DriftSchedule
	}

	return obj, nil
}

// ResolveSandboxType defaults to terraform and enforces per-type required fields.
func ResolveSandboxType(sandboxType, terraformVersion, runtime string) (string, error) {
	if sandboxType == "" {
		sandboxType = config.AppSandboxTypeTerraform
	}

	switch sandboxType {
	case config.AppSandboxTypeTerraform:
		if terraformVersion == "" {
			return "", errors.New("terraform_version is required when type=terraform")
		}
	case config.AppSandboxTypePulumi:
		if runtime == "" {
			return "", errors.New("runtime is required when type=pulumi")
		}
		if !slices.Contains(config.ValidPulumiRuntimes, runtime) {
			return "", fmt.Errorf("invalid pulumi runtime: %s. Valid runtimes: %v", runtime, config.ValidPulumiRuntimes)
		}
	default:
		return "", fmt.Errorf("invalid sandbox type: %s. Valid types: terraform, pulumi", sandboxType)
	}

	return sandboxType, nil
}

// ValidateOperationRoles fails a typo at sync rather than never matching a role.
func ValidateOperationRoles(roles []config.EntityOperationRole) error {
	for _, role := range roles {
		if !slices.Contains(app.ValidOperations, app.OperationType(role.Operation)) {
			return fmt.Errorf("invalid operation type: %s. Valid operations: %v", role.Operation, app.ValidOperations)
		}
	}
	return nil
}

func OperationRoles(roles []config.EntityOperationRole) pgtype.Hstore {
	if len(roles) == 0 {
		return nil
	}
	out := make(pgtype.Hstore, len(roles))
	for _, role := range roles {
		name := role.RoleName
		out[string(role.Operation)] = &name
	}
	return out
}

func Hstore(in map[string]string) pgtype.Hstore {
	out := make(pgtype.Hstore, len(in))
	for k, v := range in {
		val := v
		out[k] = &val
	}
	return out
}

func DerefMap(in map[string]*string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		if v == nil {
			continue
		}
		out[k] = *v
	}
	return out
}

func sliceOrEmpty(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}
