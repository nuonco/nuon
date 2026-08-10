package schema

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
)

var SchemaMapping = map[string]func() (*jsonschema.Schema, error){
	// please maintain a lexographical order here
	"action":              ActionConfigSchema,
	"branch":              AppBranchConfigSchema,
	"break-glass":         BreakGlassConfigSchema,
	"container-image":     ContainerImageConfigSchema,
	"docker-build":        DockerBuildConfigSchema,
	"helm":                HelmConfigSchema,
	"input":               InputSchema,
	"input-group":         InputGroupSchema,
	"inputs":              InputsConfigSchema,
	"install":             InstallSchema,
	"installer":           InstallerConfigSchema,
	"installs-config":     InstallsConfigSchema,
	"job":                 JobConfigSchema,
	"kubernetes-context":  KubernetesContextSchema,
	"kubernetes-contexts": KubernetesContextsConfigSchema,
	"kubernetes-manifest": KubernetesManifestConfigSchema,
	"metadata":            MetadataConfigSchema,
	"permission":          PermissionSchema,
	"permissions":         PermissionsConfigSchema,
	"policy":              PolicyConfigSchema,
	"policies":            PoliciesConfigSchema,
	"pulumi":              PulumiConfigSchema,
	"runbook":             RunbookConfigSchema,
	"runner":              RunnerConfigSchema,
	"sandbox":             SandboxConfigSchema,
	"secret":              SecretConfigSchema,
	"secrets":             SecretsConfigSchema,
	"stack":               StackConfigSchema,
	"terraform":           TerraformModuleConfigSchema,
}

// normalizeSchemaType maps underscore-style names used in TOML component
// types (e.g. docker_build, container_image) onto the hyphenated schema keys.
func normalizeSchemaType(typ string) string {
	return strings.ReplaceAll(typ, "_", "-")
}

func LookupSchemaType(typ string) (*jsonschema.Schema, error) {
	fn, ok := SchemaMapping[normalizeSchemaType(typ)]
	if !ok {
		return nil, nil
	}

	schema, err := fn()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get schema")
	}

	return schema, nil
}

// GetSchemaTypes returns all valid schema type names
func GetSchemaTypes() []string {
	types := make([]string, 0, len(SchemaMapping))
	for k := range SchemaMapping {
		types = append(types, k)
	}
	return types
}

// IsValidSchemaType checks if a schema type is valid
func IsValidSchemaType(typ string) bool {
	_, ok := SchemaMapping[normalizeSchemaType(typ)]
	return ok
}

// Schema functions in lexicographical order

func ActionConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.ActionConfig{}); err != nil {
		return nil, errors.Wrap(err, "ActionConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.ActionConfig{}), nil
}

func AppBranchConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppBranchConfig{}); err != nil {
		return nil, errors.Wrap(err, "AppBranchConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppBranchConfig{}), nil
}

func AppConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppConfig{}); err != nil {
		return nil, errors.Wrap(err, "AppConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppConfig{}), nil
}

func BreakGlassConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.BreakGlass{}); err != nil {
		return nil, errors.Wrap(err, "BreakGlass validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.BreakGlass{}), nil
}

func InputGroupSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppInputGroup{}); err != nil {
		return nil, errors.Wrap(err, "AppInputGroup validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppInputGroup{}), nil
}

func InputSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppInput{}); err != nil {
		return nil, errors.Wrap(err, "AppInput validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppInput{}), nil
}

func InputsConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppInputConfig{}); err != nil {
		return nil, errors.Wrap(err, "AppInputConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppInputConfig{}), nil
}

func InstallSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.Install{}); err != nil {
		return nil, errors.Wrap(err, "Install validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.Install{}), nil
}

func InstallerConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.InstallerConfig{}); err != nil {
		return nil, errors.Wrap(err, "InstallerConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.InstallerConfig{}), nil
}

func InstallsConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.InstallsConfig{}); err != nil {
		return nil, errors.Wrap(err, "InstallsConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.InstallsConfig{}), nil
}

func KubernetesContextSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.KubernetesContext{}); err != nil {
		return nil, errors.Wrap(err, "KubernetesContext validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.KubernetesContext{}), nil
}

func KubernetesContextsConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.KubernetesContextsConfig{}); err != nil {
		return nil, errors.Wrap(err, "KubernetesContextsConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.KubernetesContextsConfig{}), nil
}

func KubernetesManifestConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.KubernetesManifestComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "KubernetesManifestComponentConfig validation failed")
	}

	return flattenedComponentSchema(config.KubernetesManifestComponentConfig{}, config.KubernetesManifestComponentType)
}

func ContainerImageConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.ExternalImageComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "ExternalImageComponentConfig validation failed")
	}

	return flattenedComponentSchema(config.ExternalImageComponentConfig{}, config.ContainerImageComponentType, config.ExternalImageComponentType)
}

func DockerBuildConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.DockerBuildComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "DockerBuildComponentConfig validation failed")
	}

	return flattenedComponentSchema(config.DockerBuildComponentConfig{}, config.DockerBuildComponentType)
}

func JobConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.JobComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "JobComponentConfig validation failed")
	}

	return flattenedComponentSchema(config.JobComponentConfig{}, config.JobComponentType)
}

func HelmConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.HelmChartComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "HelmChartComponentConfig validation failed")
	}

	return flattenedComponentSchema(config.HelmChartComponentConfig{}, config.HelmChartComponentType)
}

func TerraformModuleConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.TerraformModuleComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "TerraformModuleComponentConfig validation failed")
	}

	return flattenedComponentSchema(config.TerraformModuleComponentConfig{}, config.TerraformModuleComponentType)
}

func PulumiConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.PulumiComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "PulumiComponentConfig validation failed")
	}

	return flattenedComponentSchema(config.PulumiComponentConfig{}, config.PulumiComponentType)
}

func MetadataConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.MetadataConfig{}); err != nil {
		return nil, errors.Wrap(err, "MetadataConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.MetadataConfig{}), nil
}

// PermissionSchema is the schema for a single IAM role file, as used by the
// permissions/ and break_glass/ directory forms (each file is one
// AppAWSIAMRole).
func PermissionSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppAWSIAMRole{}); err != nil {
		return nil, errors.Wrap(err, "AppAWSIAMRole validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppAWSIAMRole{}), nil
}

func PermissionsConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.PermissionsConfig{}); err != nil {
		return nil, errors.Wrap(err, "PermissionsConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.PermissionsConfig{}), nil
}

func PolicyConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppPolicy{}); err != nil {
		return nil, errors.Wrap(err, "AppPolicy validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppPolicy{}), nil
}

func PoliciesConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.PoliciesConfig{}); err != nil {
		return nil, errors.Wrap(err, "PoliciesConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.PoliciesConfig{}), nil
}

func RunbookConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.RunbookConfig{}); err != nil {
		return nil, errors.Wrap(err, "RunbookConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.RunbookConfig{}), nil
}

func RunnerConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppRunnerConfig{}); err != nil {
		return nil, errors.Wrap(err, "AppRunnerConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppRunnerConfig{}), nil
}

func SandboxConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppSandboxConfig{}); err != nil {
		return nil, errors.Wrap(err, "AppSandboxConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppSandboxConfig{}), nil
}

func SecretConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.AppSecret{}); err != nil {
		return nil, errors.Wrap(err, "AppSecret validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppSecret{}), nil
}

func SecretsConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.SecretsConfig{}); err != nil {
		return nil, errors.Wrap(err, "SecretsConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.SecretsConfig{}), nil
}

func StackConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.StackConfig{}); err != nil {
		return nil, errors.Wrap(err, "StackConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.StackConfig{}), nil
}

// ValidateJSONSchemaExtend checks that a struct and all its nested struct fields
// implement JSONSchemaExtend before being passed to reflection.
// This ensures proper schema generation with custom extensions.
func ValidateJSONSchemaExtend(structVal interface{}) error {
	return validateStructHasJSONSchemaExtend(reflect.TypeOf(structVal), "")
}

func validateStructHasJSONSchemaExtend(t reflect.Type, fieldPath string) error {
	// Dereference pointers
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle slices and arrays
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		return validateStructHasJSONSchemaExtend(t.Elem(), fieldPath)
	}

	// Skip validation for jsonschema.Schema type (built-in exception)
	if t == reflect.TypeOf(jsonschema.Schema{}) {
		return nil
	}

	// Only validate struct types
	if t.Kind() != reflect.Struct {
		return nil
	}

	// Check if this struct implements JSONSchemaExtend
	method, ok := t.MethodByName("JSONSchemaExtend")
	if !ok || method.Type.NumIn() != 2 || method.Type.In(1).String() != "*jsonschema.Schema" {
		fullPath := fieldPath
		if fullPath == "" {
			fullPath = t.Name()
		}
		return fmt.Errorf("struct %s does not implement JSONSchemaExtend(*jsonschema.Schema)", fullPath)
	}

	// Validate nested struct fields
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip fields with jsonschema:"-" tag
		if field.Tag.Get("jsonschema") == "-" {
			continue
		}

		fieldType := field.Type
		// Dereference pointers for field type checking
		for fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Check slices/arrays
		if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
			fieldType = fieldType.Elem()
			// Dereference again if pointer element
			for fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
		}

		// If it's a struct, validate it recursively
		if fieldType.Kind() == reflect.Struct {
			nestedPath := fieldPath
			if nestedPath != "" {
				nestedPath += "." + field.Name
			} else {
				nestedPath = t.Name() + "." + field.Name
			}

			if err := validateStructHasJSONSchemaExtend(fieldType, nestedPath); err != nil {
				return err
			}
		}
	}

	return nil
}
