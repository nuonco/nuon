package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/schema"
)

const schemaBaseURL = "https://api.nuon.co"

// schemaTypeForDefinition resolves the schema type slug for a config file. It
// prefers an explicitly-set Header, then falls back to deriving the slug from
// the config instances the file was built from, so scaffolded files that don't
// set Header still get a #:schema directive.
func schemaTypeForDefinition(cfd ConfigFileDefinition) string {
	if cfd.Header != "" {
		return cfd.Header
	}

	for _, cfs := range cfd.Schemas {
		if slug := schemaTypeForInstance(cfs.Instance); slug != "" {
			return slug
		}
	}

	return ""
}

func schemaTypeForInstance(instance any) string {
	switch v := instance.(type) {
	case *config.Component:
		return componentTypeSchemaSlug(v.Type)
	case config.Component:
		return componentTypeSchemaSlug(v.Type)
	case *config.AppInputConfig:
		return "inputs"
	case *config.InstallerConfig:
		return "installer"
	case *config.AppSandboxConfig:
		return "sandbox"
	case *config.AppRunnerConfig:
		return "runner"
	case *config.StackConfig:
		return "stack"
	case *config.SecretsConfig:
		return "secrets"
	case *config.BreakGlass:
		return "break-glass"
	case *config.PoliciesConfig:
		return "policies"
	case *config.PermissionsConfig:
		return "permissions"
	case *config.ActionConfig:
		return "action"
	case *config.Install:
		return "install"
	case *config.MetadataConfig:
		return "metadata"
	case *config.TerraformModuleComponentConfig:
		return "terraform"
	case *config.HelmChartComponentConfig:
		return "helm"
	case *config.DockerBuildComponentConfig:
		return "docker-build"
	case *config.ExternalImageComponentConfig:
		return "container-image"
	case *config.KubernetesManifestComponentConfig:
		return "kubernetes-manifest"
	case *config.JobComponentConfig:
		return "job"
	default:
		return ""
	}
}

func componentTypeSchemaSlug(t config.ComponentType) string {
	switch t {
	case config.TerraformModuleComponentType:
		return "terraform"
	case config.HelmChartComponentType:
		return "helm"
	case config.DockerBuildComponentType:
		return "docker-build"
	case config.ContainerImageComponentType, config.ExternalImageComponentType:
		return "container-image"
	case config.KubernetesManifestComponentType:
		return "kubernetes-manifest"
	case config.JobComponentType:
		return "job"
	default:
		return ""
	}
}

func NewDefaultReflector() *jsonschema.Reflector {
	return &jsonschema.Reflector{
		ExpandedStruct:            true,
		Anonymous:                 true,
		FieldNameTag:              "mapstructure",
		DoNotReference:            true,
		AllowAdditionalProperties: false,
	}
}

const (
	StructTagOneofRequired                   = "oneof_required"
	StructTagOneofRequiredGroupComponentType = "component_type"
	StructTagOneofRequiredGroupGitRepository = "git_repository"
)

var (
	StructTagOneOfRequiredGroups = []string{StructTagOneofRequiredGroupComponentType, StructTagOneofRequiredGroupGitRepository}
	IgnoredProperties            = []string{
		"source",
		"helm_chart",
		"terraform_module",
		"docker_build",
		"job",
		"external_image",
		"kubernetes_manifest",
	}
)

type ConfigGen struct {
	EnableDefaults          bool
	EnableInfoComments      bool
	EnableDeprecated        bool
	SkipNonRequired         bool
	OverwriteConfigContents bool
}

func NewConfigGen(EnableDefaults, EnableInfoComments, EnableDeprecated, OverwriteConfigContents, SkipNonRequired bool) *ConfigGen {
	return &ConfigGen{
		EnableDefaults:          EnableDefaults,
		EnableInfoComments:      EnableInfoComments,
		EnableDeprecated:        EnableDeprecated,
		OverwriteConfigContents: OverwriteConfigContents,
		SkipNonRequired:         SkipNonRequired,
	}
}

func (g *ConfigGen) Validate(path string) error {
	// path needs to be a directory not a file
	stat, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	// if path doesn't exist, it's valid (we can create it later)
	if os.IsNotExist(err) {
		return nil
	}

	// path exists but is not a directory, error out
	if stat != nil && !stat.IsDir() {
		return fmt.Errorf("path %s is not a directory", path)
	}

	// if directory exists, check if it's empty
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}
	if len(entries) > 0 && !g.OverwriteConfigContents {
		return fmt.Errorf("directory %s is not empty", path)
	}

	return nil
}

func (g *ConfigGen) Gen(path string, c *ConfigStructure) error {
	if err := g.Validate(path); err != nil {
		return errors.Wrap(err, "input validation error")
	}

	if c == nil {
		c = DefaultAppConfigConfigStructure(path)
	}

	if c.Name == "" {
		c.Name = path
	}

	err := g.EncodeToTOML(c)
	if err != nil {
		return errors.Wrapf(err, "unable to encode to TOML")
	}

	err = g.WriteConfigToDisk(c)
	if err != nil {
		return errors.Wrapf(err, "unable to write config to disk")
	}
	return nil
}

func (g *ConfigGen) WriteConfigToDisk(c *ConfigStructure) error {
	if _, err := os.Stat(c.Name); err != nil && os.IsNotExist(err) {
		// create config directory if it doesn't exist
		if err := os.Mkdir(c.Name, 0o755); err != nil {
			return errors.Wrapf(err, "unable to create app config directory for path: %s", c.Name)
		}
	}

	for _, f := range c.Configs {
		fp := filepath.Join(c.Name)
		fp = filepath.Join(fp, f.Name)
		if err := os.WriteFile(fp, []byte(strings.TrimSpace(f.TomlEncoded)), 0o644); err != nil {
			return errors.Wrapf(err, "failed to write schema file %s", fp)
		}
	}

	for _, d := range c.ConfigDirectories {
		dp := filepath.Join(c.Name)
		dp = filepath.Join(dp, d.Name)

		if err := os.Mkdir(dp, 0o755); err != nil && !os.IsExist(err) {
			return errors.Wrapf(err, "unable to create app config sub-dirctory for path: %s", dp)
		}

		for _, f := range d.Configs {
			fp := filepath.Join(dp, f.Name)

			if err := os.WriteFile(fp, []byte(strings.TrimSpace(f.TomlEncoded)), 0o644); err != nil {
				return errors.Wrapf(err, "failed to write schema file : %s", fp)
			}
		}
	}

	return nil
}

func (g *ConfigGen) EncodeToTOML(cs *ConfigStructure) error {
	for fi, f := range cs.Configs {
		tomlEncoded, err := g.encodeConfigFile(f, f.Name)
		if err != nil {
			return err
		}
		cs.Configs[fi].TomlEncoded = tomlEncoded.String()
	}
	for di, d := range cs.ConfigDirectories {
		for fi, f := range d.Configs {
			tomlEncoded, err := g.encodeConfigFile(f, d.Name+"/"+f.Name)
			if err != nil {
				return err
			}
			cs.ConfigDirectories[di].Configs[fi].TomlEncoded = tomlEncoded.String()
		}
	}
	return nil
}

// encodeConfigFile returns contents of a file
func (g *ConfigGen) encodeConfigFile(cfd ConfigFileDefinition, name string) (*strings.Builder, error) {
	var output strings.Builder

	// write the schema directive so editors with a TOML LSP resolve the
	// dedicated per-type schema for this file
	if slug := schemaTypeForDefinition(cfd); schema.IsValidSchemaType(slug) {
		output.WriteString(fmt.Sprintf("#:schema %s/v1/general/config-schema/%s\n\n", schemaBaseURL, slug))
	}

	for _, configFile := range cfd.Schemas {
		schema := configFile.Schema()
		if schema == nil {
			continue
		}

		extractor := NewInstanceValueExtractor(configFile.Instance)

		oneOFGroups := make(map[string]map[string]bool)
		for _, s := range schema.OneOf {
			oneOFGroups[s.Title] = make(map[string]bool)
			oneOfRequired := oneOFGroups[s.Title]
			for _, r := range s.Required {
				oneOfRequired[r] = true
			}
		}

		g.recursivelyEncode(schema, oneOFGroups, &output, "", false, g.EnableInfoComments, configFile.SkipNonRequired, extractor)
	}
	return &output, nil
}
