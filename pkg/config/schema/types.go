package schema

import (
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config"
)

var SchemaMapping = map[string]func() (*jsonschema.Schema, error){
	// please maintain a lexographical order here
	"action":              ActionConfigSchema,
	"break-glass":         BreakGlassConfigSchema,
	"component-root":      ComponentRootSchema,
	"container-image":     ContainerImageConfigSchema,
	"docker-build":        DockerBuildConfigSchema,
	"full":                AppConfigSchema,
	"helm":                HelmConfigSchema,
	"input":               InputSchema,
	"input-group":         InputGroupSchema,
	"inputs":              InputsConfigSchema,
	"install":             InstallSchema,
	"installer":           InstallerConfigSchema,
	"kubernetes-manifest": KubernetesManifestConfigSchema,
	"metadata":            MetadataConfigSchema,
	"permissions":         PermissionsConfigSchema,
	"policy":              PolicyConfigSchema,
	"policies":            PoliciesConfigSchema,
	"runner":              RunnerConfigSchema,
	"sandbox":             SandboxConfigSchema,
	"secret":              SecretConfigSchema,
	"secrets":             SecretsConfigSchema,
	"stack":               StackConfigSchema,
	"terraform":           TerraformModuleConfigSchema,
}

func LookupSchemaType(typ string) (*jsonschema.Schema, error) {
	fn, ok := SchemaMapping[typ]
	if !ok {
		return nil, nil
	}

	schema, err := fn()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get schema")
	}

	return schema, nil
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
	if err := ValidateJSONSchemaExtend(config.AppAWSIAMRole{}); err != nil {
		return nil, errors.Wrap(err, "AppAWSIAMRole validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.AppAWSIAMRole{}), nil
}

func ComponentRootSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.Component{}); err != nil {
		return nil, errors.Wrap(err, "Component validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.Component{}), nil
}

func ContainerImageConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.ExternalImageComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "ExternalImageComponentConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.ExternalImageComponentConfig{}), nil
}

func DockerBuildConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.DockerBuildComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "DockerBuildComponentConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.DockerBuildComponentConfig{}), nil
}

func HelmConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.HelmChartComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "HelmChartComponentConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.HelmChartComponentConfig{}), nil
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

func KubernetesManifestConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.KubernetesManifestComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "KubernetesManifestComponentConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.KubernetesManifestComponentConfig{}), nil
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

func TerraformModuleConfigSchema() (*jsonschema.Schema, error) {
	if err := ValidateJSONSchemaExtend(config.TerraformModuleComponentConfig{}); err != nil {
		return nil, errors.Wrap(err, "TerraformModuleComponentConfig validation failed")
	}

	r, err := reflector()
	if err != nil {
		return nil, err
	}

	return r.Reflect(config.TerraformModuleComponentConfig{}), nil
}

// ValidateJSONSchemaExtend checks that a struct and all its nested struct fields
// implement JSONSchemaExtend before being passed to reflection.
// This ensures proper schema generation with custom extensions.
func ValidateJSONSchemaExtend(structVal interface{}) error {
	return validateStructHasJSONSchemaExtend(reflect.TypeOf(structVal), "")
}

//nolint:gocyclo
func validateStructHasJSONSchemaExtend(t reflect.Type, fieldPath string) error {
	// Dereference pointers
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle slices and arrays
	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		return validateStructHasJSONSchemaExtend(t.Elem(), fieldPath)
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
