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

const defaultSchemaBaseURL = "https://api.nuon.co"

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

var IgnoredProperties = []string{
	"source",
	"helm_chart",
	"terraform_module",
	"docker_build",
	"job",
	"external_image",
	"kubernetes_manifest",
	"pulumi",
	"slack_webhook_url",
}

type ConfigGen struct {
	EnableDefaults          bool
	EnableInfoComments      bool
	EnableDeprecated        bool
	SkipNonRequired         bool
	OverwriteConfigContents bool

	SchemaBaseURL string
}

func (g *ConfigGen) schemaBaseURL() string {
	if g.SchemaBaseURL != "" {
		return strings.TrimSuffix(g.SchemaBaseURL, "/")
	}
	return defaultSchemaBaseURL
}

func NewConfigGen(EnableDefaults, EnableInfoComments, EnableDeprecated, OverwriteConfigContents, SkipNonRequired bool, schemaBaseURL string) *ConfigGen {
	return &ConfigGen{
		EnableDefaults:          EnableDefaults,
		EnableInfoComments:      EnableInfoComments,
		EnableDeprecated:        EnableDeprecated,
		OverwriteConfigContents: OverwriteConfigContents,
		SkipNonRequired:         SkipNonRequired,
		SchemaBaseURL:           schemaBaseURL,
	}
}

func (g *ConfigGen) Validate(path string) error {
	stat, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if os.IsNotExist(err) {
		return nil
	}

	if stat != nil && !stat.IsDir() {
		return fmt.Errorf("path %s is not a directory", path)
	}

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

	for _, f := range c.RawFiles {
		fp := filepath.Join(c.Name, f.Name)
		if err := os.WriteFile(fp, []byte(f.Contents), 0o644); err != nil {
			return errors.Wrapf(err, "failed to write file %s", fp)
		}
	}

	for _, d := range c.ConfigDirectories {
		dp := filepath.Join(c.Name)
		dp = filepath.Join(dp, d.Name)

		if err := os.MkdirAll(dp, 0o755); err != nil && !os.IsExist(err) {
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

func alignAssignments(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))

	var group []int
	flush := func() {
		if len(group) < 1 {
			group = nil
			return
		}
		width := 0
		for _, idx := range group {
			key := strings.TrimRight(out[idx][:strings.Index(out[idx], "=")], " ")
			if len(key) > width {
				width = len(key)
			}
		}
		for _, idx := range group {
			eq := strings.Index(out[idx], "=")
			key := strings.TrimRight(out[idx][:eq], " ")
			val := strings.TrimLeft(out[idx][eq+1:], " ")
			out[idx] = fmt.Sprintf("%-*s = %s", width, key, val)
		}
		group = nil
	}

	inMultiline := false
	for _, line := range lines {
		if inMultiline {
			out = append(out, line)
			if strings.Count(line, `"""`)%2 == 1 {
				inMultiline = false
			}
			continue
		}

		if strings.TrimSpace(line) == "" {
			flush()
			if len(out) > 0 && out[len(out)-1] == "" {
				continue
			}
			out = append(out, "")
			continue
		}

		if strings.Count(line, `"""`)%2 == 1 {
			flush()
			out = append(out, line)
			inMultiline = true
			continue
		}

		if isAssignmentLine(line) {
			group = append(group, len(out))
			out = append(out, line)
			continue
		}

		flush()
		if isTableHeader(line) && len(out) > 0 && out[len(out)-1] != "" {
			out = append(out, "")
		}
		out = append(out, line)
	}
	flush()

	result := strings.Join(out, "\n")
	return strings.TrimRight(result, "\n") + "\n"
}

func isAssignmentLine(line string) bool {
	core := strings.TrimSpace(line)
	if strings.HasPrefix(core, "#") {
		core = strings.TrimSpace(strings.TrimPrefix(core, "#"))
	}
	if core == "" || strings.HasPrefix(core, "[") || strings.HasPrefix(core, ":") {
		return false
	}
	eq := strings.Index(core, "=")
	return eq > 0
}

func isTableHeader(line string) bool {
	core := strings.TrimSpace(line)
	if strings.HasPrefix(core, "#") {
		core = strings.TrimSpace(strings.TrimPrefix(core, "#"))
	}
	return strings.HasPrefix(core, "[")
}

func (g *ConfigGen) encodeConfigFile(cfd ConfigFileDefinition, name string) (*strings.Builder, error) {
	var output strings.Builder

	if slug := schemaTypeForDefinition(cfd); schema.IsValidSchemaType(slug) {
		output.WriteString(fmt.Sprintf("#:schema %s/v1/general/config-schema/%s\n\n", g.schemaBaseURL(), slug))
	}

	for _, phase := range []encodePhase{phaseLines, phaseBlocks} {
		for _, configFile := range cfd.Schemas {
			schema := configFile.Schema()
			if schema == nil {
				continue
			}

			extractor := NewInstanceValueExtractor(configFile.Instance)

			g.recursivelyEncode(schema, &output, "", false, g.EnableInfoComments, configFile.SkipNonRequired, extractor, phase, false)
		}
	}

	var formatted strings.Builder
	formatted.WriteString(alignAssignments(output.String()))
	return &formatted, nil
}
