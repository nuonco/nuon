package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	configparse "github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/config/parse/get"
	"github.com/nuonco/nuon/pkg/rootfs"
	"github.com/nuonco/nuon/pkg/sourcefs"

	"github.com/go-playground/validator/v10"
)

type ParseInstallConfigsInput struct {
	SourceDir         string   `json:"source_dir"`
	InstallsDirectory string   `json:"installs_directory"`
	InstallName       string   `json:"install_name,omitempty"`
	ChangedFiles      []string `json:"changed_files,omitempty"`
}

type ParsedInstallConfig struct {
	Config   *config.Install `json:"config"`
	FilePath string          `json:"file_path"`
}

type ParseInstallConfigsResponse struct {
	Installs []ParsedInstallConfig `json:"installs"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
// @by-field sourceDir
func (a *Activities) parseInstallConfigs(ctx context.Context, sourceDir string, req *ParseInstallConfigsInput) (*ParseInstallConfigsResponse, error) {
	installsPath := filepath.Clean(strings.TrimLeft(req.InstallsDirectory, "/"))
	if !filepath.IsLocal(installsPath) || strings.Contains(req.InstallsDirectory, `\`) {
		return nil, fmt.Errorf("invalid installs directory %q", req.InstallsDirectory)
	}
	root, err := rootfs.Open(sourceDir)
	if os.IsNotExist(err) {
		return &ParseInstallConfigsResponse{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to open source directory: %w", err)
	}
	defer root.Close()
	source, err := sourcefs.From(root, root.Root())
	if err != nil {
		return nil, err
	}

	resolved, err := source.Resolve(installsPath)
	if os.IsNotExist(err) {
		return &ParseInstallConfigsResponse{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to resolve installs directory: %w", err)
	}
	info, err := source.Lstat(resolved)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return &ParseInstallConfigsResponse{}, nil
	}

	if req.InstallName != "" {
		return a.parseSingleInstall(ctx, source, installsPath, req.InstallName)
	}

	changedSet := make(map[string]bool, len(req.ChangedFiles))
	for _, f := range req.ChangedFiles {
		changedSet[f] = true
	}

	var result []ParsedInstallConfig

	entries, err := source.ReadDir(installsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read installs directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		relPath := filepath.Join(req.InstallsDirectory, entry.Name())

		if len(changedSet) > 0 && !changedSet[relPath] {
			continue
		}

		fullPath := filepath.Join(installsPath, entry.Name())
		installCfg, err := parseInstallTOML(ctx, source, fullPath)
		if err != nil {
			return nil, fmt.Errorf("unable to parse %s: %w", relPath, err)
		}

		result = append(result, ParsedInstallConfig{
			Config:   installCfg,
			FilePath: relPath,
		})
	}

	return &ParseInstallConfigsResponse{Installs: result}, nil
}

func (a *Activities) parseSingleInstall(ctx context.Context, source *sourcefs.FS, installsPath, installName string) (*ParseInstallConfigsResponse, error) {
	if installName == "." || installName == ".." || strings.ContainsAny(installName, `/\`) {
		return nil, fmt.Errorf("invalid install name %q", installName)
	}
	fileName := installName + ".toml"
	fullPath := filepath.Join(installsPath, fileName)

	_, err := source.Resolve(fullPath)
	if os.IsNotExist(err) {
		return &ParseInstallConfigsResponse{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to resolve %s: %w", fileName, err)
	}
	installCfg, err := parseInstallTOML(ctx, source, fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s: %w", fileName, err)
	}

	return &ParseInstallConfigsResponse{
		Installs: []ParsedInstallConfig{
			{Config: installCfg, FilePath: fileName},
		},
	}, nil
}

func parseInstallTOML(ctx context.Context, source *sourcefs.FS, filePath string) (*config.Install, error) {
	f, err := source.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %w", filePath, err)
	}
	defer f.Close()

	cfg, err := configparse.ParseInstallConfig(f, validator.New())
	if err != nil {
		return nil, fmt.Errorf("unable to parse install config: %w", err)
	}
	if err := configparse.LoadInstallUIDefinition(ctx, cfg, &get.Options{
		RootDir: filepath.Join(source.Root(), filepath.Dir(filePath)),
		Source:  source,
	}); err != nil {
		return nil, err
	}

	return cfg, nil
}
