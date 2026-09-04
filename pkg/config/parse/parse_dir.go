package parse

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse/dir"
	"github.com/nuonco/nuon/pkg/config/parse/get"
)

const (
	defaultFieldGetTimeout time.Duration = time.Second * 5
)

func parseDirName(dirname string) (string, error) {
	if _, err := os.Stat(dirname); err != nil {
		if os.IsNotExist(err) {
			return "", config.ErrConfig{
				Description: dirname + " does not exist",
				Err:         err,
			}
		}

		return "", err
	}

	absPath, err := filepath.Abs(dirname)
	if err != nil {
		return "", errors.Wrap(err, "unable to get absolute path")
	}

	return absPath, nil
}

func ParseDir(ctx context.Context, parseCfg ParseConfig) (*config.AppConfig, error) {
	result, err := parseDir(ctx, parseCfg, nil)
	if err != nil {
		return nil, err
	}
	return result.Config, nil
}

type ParseResult struct {
	Config *config.AppConfig
	Source *config.SourceArchive
}

func ParseDirWithSource(ctx context.Context, parseCfg ParseConfig) (*ParseResult, error) {
	return parseDir(ctx, parseCfg, &sourceCapture{archive: config.NewSourceArchive()})
}

type sourceCapture struct {
	archive *config.SourceArchive
	err     error
}

func (s *sourceCapture) recordError(err error) {
	if err != nil && s.err == nil {
		s.err = err
	}
}

func parseDir(ctx context.Context, parseCfg ParseConfig, source *sourceCapture) (*ParseResult, error) {
	fp, err := parseDirName(parseCfg.Dirname)
	if err != nil {
		return nil, err
	}

	fs := afero.NewOsFs()
	cfgFS := afero.NewBasePathFs(fs, fp)

	// parse the directory
	var obj ConfigDir
	if err := dir.Parse(ctx, cfgFS, &obj, &dir.ParseOptions{
		Root:         fp,
		Ext:          ".toml",
		ParserFn:     func(rc io.ReadCloser, s string, a any) error { return parseTomlFile(rc, s, a, parseCfg.FileProcessor) },
		OnParsedFile: sourceFileRecorder(source),
	}); err != nil {
		return nil, errors.Wrap(err, "unable to parse directory")
	}

	hasConfigs, err := hasTomlFiles(cfgFS)
	if err != nil {
		return nil, ParseErr{
			Description: "error checking for toml files in directory",
			Err:         err,
		}
	}
	if !hasConfigs {
		return nil, ParseErr{
			Description: "no configuration files found in directory",
			Err:         fmt.Errorf("no configuration files found in directory %s", fp),
		}
	}

	// NOTE(jm): this will go away once we deprecate the legacy config, and we can just have a pipeline of
	// `config.AppConfig` parsers.
	appCfg, err := obj.toAppConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert to app config")
	}
	if source != nil && source.err != nil {
		return nil, ParseErr{
			Description: "unable to archive authored config",
			Err:         source.err,
		}
	}

	// Derive policy names from Contents paths BEFORE get.Parse() replaces Contents with actual file content.
	// This allows policies defined in policies.toml with Contents like "./block-mutable-tags.rego" to derive
	// their name as "block-mutable-tags" before the path is replaced with the file content.
	if appCfg.Policies != nil {
		for i := range appCfg.Policies.Policies {
			appCfg.Policies.Policies[i].SetNameFromContents()
		}
	}

	// Copy TemplateURL to Contents for go-getter to fetch template content.
	if appCfg.Stack != nil {
		for i := range appCfg.Stack.CustomNestedStacks {
			appCfg.Stack.CustomNestedStacks[i].Contents = appCfg.Stack.CustomNestedStacks[i].TemplateURL
		}
	}

	fieldTimeout := defaultFieldGetTimeout
	if parseCfg.FieldTimeout > 0 {
		fieldTimeout = parseCfg.FieldTimeout
	}

	// parse all get functions
	if err := get.Parse(ctx, appCfg, &get.Options{
		FieldTimeout: fieldTimeout,
		RootDir:      fp,
		OnLocalFile:  localSourceFileRecorder(source),
	}); err != nil {
		return nil, ParseErr{
			Description: "unable to get fields",
			Err:         err,
		}
	}

	err = appCfg.Parse()
	if err != nil {
		return nil, ParseErr{
			Description: "error parsing config",
			Err:         err,
		}
	}

	checksums, err := checksumTOMLFilesByName(cfgFS)
	if err != nil {
		return nil, errors.Wrap(err, "unable to checksum toml files")
	}

	for _, cmp := range appCfg.Components {
		if cmp == nil {
			continue
		}

		if checksum, ok := checksums[cmp.Name]; ok {
			cmp.Checksum = checksum
		}
	}

	if source == nil {
		return &ParseResult{Config: appCfg}, nil
	}
	source.recordError(source.archive.ReindexMembers())
	if source.err != nil {
		return nil, ParseErr{
			Description: "unable to archive authored config",
			Err:         source.err,
		}
	}
	return &ParseResult{Config: appCfg, Source: source.archive}, nil
}

func sourceFileRecorder(source *sourceCapture) func(dir.ParsedFile) error {
	if source == nil {
		return nil
	}
	return func(file dir.ParsedFile) error {
		if err := source.archive.AddFile(file.Path, file.Contents); err != nil {
			source.recordError(err)
			return nil
		}
		if _, ok := source.archive.Files[filepath.ToSlash(file.Path)]; !ok {
			return nil
		}
		kind, name := sourceMemberIdentity(file.Group, file.Path, file.Value)
		if kind != "" && name == "" {
			source.recordError(fmt.Errorf("source file %s has no name field", file.Path))
			return nil
		}
		source.recordError(source.archive.AddMember(kind, name, file.Path))
		return nil
	}
}

func localSourceFileRecorder(source *sourceCapture) func(string, []byte) error {
	if source == nil {
		return nil
	}
	return func(path string, contents []byte) error {
		source.recordError(source.archive.AddFile(path, contents))
		return nil
	}
}

func sourceMemberIdentity(group, path string, value any) (string, string) {
	if path == group+".toml" {
		switch group {
		case "metadata":
			return "metadata", "metadata"
		case "inputs":
			return "input", "inputs"
		case "policies":
			return "policy", "policies"
		case "break_glass":
			return "break_glass", "break-glass"
		case "sandbox", "stack", "runner":
			return group, group
		}
	}

	kind := ""
	switch group {
	case "components":
		kind = "component"
	case "actions":
		kind = "action"
	case "runbooks":
		kind = "runbook"
	case "permissions":
		kind = "permission"
	default:
		return "", ""
	}

	reflected := reflect.ValueOf(value)
	for reflected.IsValid() && reflected.Kind() == reflect.Ptr {
		if reflected.IsNil() {
			return "", ""
		}
		reflected = reflected.Elem()
	}
	if !reflected.IsValid() || reflected.Kind() != reflect.Struct {
		return "", ""
	}
	field := reflected.FieldByName("Name")
	if !field.IsValid() || field.Kind() != reflect.String {
		return "", ""
	}
	return kind, field.String()
}

func hasTomlFiles(fs afero.Fs) (bool, error) {
	// Read directory contents
	files, err := afero.ReadDir(fs, ".")
	if err != nil {
		return false, err
	}

	// Check each file for .toml extension
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".toml") {
			return true, nil
		}
	}

	return false, nil
}

// Deprecated: we not hash the intermediatery strucst for component configs
func checksumTOMLFilesByName(cfgFS afero.Fs) (map[string]string, error) {
	checksums := make(map[string]string)

	// Read the components directory
	files, err := afero.ReadDir(cfgFS, "components")
	if err != nil {
		return nil, fmt.Errorf("failed to read components directory: %w", err)
	}

	for _, file := range files {
		// Skip directories and non-TOML files
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".toml") {
			continue
		}

		filePath := filepath.Join("components", file.Name())

		// Read file content
		content, err := afero.ReadFile(cfgFS, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		// Parse TOML to get the name
		var config config.Component
		if err := toml.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("failed to parse TOML in %s: %w", filePath, err)
		}

		// Skip files without a name field
		if config.Name == "" {
			fmt.Printf("Warning: %s has no 'name' field, skipping\n", filePath)
			continue
		}

		// Calculate SHA256 checksum
		hash := sha256.Sum256(content)
		checksums[config.Name] = fmt.Sprintf("%x", hash)
	}

	return checksums, nil
}
