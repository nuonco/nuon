package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
)

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
)

var StructTagOneOfRequiredGroups = []string{"component_type", "git_repository"}

/*
todo(sk):
- fix the structs where there are multiple oneofs eg, component, permission
- fix jsonschema where some values are not required and are commented
- fix jsonschema in root app config
*/
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

// todo: fix the folder permissions, should not be 777
func (g *ConfigGen) WriteConfigToDisk(c *ConfigStructure) error {
	if _, err := os.Stat(c.Name); err != nil && os.IsNotExist(err) {
		// create config directory if it doesn't exist
		if err := os.Mkdir(c.Name, 0o777); err != nil {
			return errors.Wrapf(err, "unable to create app config directory for path: %s", c.Name)
		}
	}

	for _, f := range c.Configs {
		fp := filepath.Join(c.Name)
		fp = filepath.Join(fp, f.Name)
		if err := os.WriteFile(fp, []byte(strings.TrimSpace(f.TomlEncoded)), 0o777); err != nil {
			return errors.Wrapf(err, "failed to write schema file %s", fp)
		}
	}

	for _, d := range c.ConfigDirectories {
		dp := filepath.Join(c.Name)
		dp = filepath.Join(dp, d.Name)

		if err := os.Mkdir(dp, 0o777); err != nil && !os.IsExist(err) {
			return errors.Wrapf(err, "unable to create app config sub-dirctory for path: %s", dp)
		}

		for _, f := range d.Configs {
			fp := filepath.Join(dp, f.Name)

			if err := os.WriteFile(fp, []byte(strings.TrimSpace(f.TomlEncoded)), 0o777); err != nil {
				return errors.Wrapf(err, "failed to write schema file : %s", fp)
			}
		}
	}

	return nil
}

func (g *ConfigGen) EncodeToTOML(cs *ConfigStructure) error {
	for fi, f := range cs.Configs {
		tomlEncoded, err := g.encodeConfigFile(f.Schemas, f.Name)
		if err != nil {
			return err
		}
		cs.Configs[fi].TomlEncoded = tomlEncoded.String()
	}
	for di, d := range cs.ConfigDirectories {
		for fi, f := range d.Configs {
			tomlEncoded, err := g.encodeConfigFile(f.Schemas, d.Name+"/"+f.Name)
			if err != nil {
				return err
			}
			// fmt.Println(tomlEncoded.String())
			cs.ConfigDirectories[di].Configs[fi].TomlEncoded = tomlEncoded.String()
		}
	}
	return nil
}

// encodeConfigFile returns contents of a file
func (g *ConfigGen) encodeConfigFile(configFileSchemas []ConfigFileSchema, name string) (*strings.Builder, error) {
	var output strings.Builder
	for _, configFile := range configFileSchemas {
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
