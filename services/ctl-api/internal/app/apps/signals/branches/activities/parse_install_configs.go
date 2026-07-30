package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	configparse "github.com/nuonco/nuon/pkg/config/parse"

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
	installsPath := filepath.Join(sourceDir, req.InstallsDirectory)

	info, err := os.Stat(installsPath)
	if err != nil || !info.IsDir() {
		return &ParseInstallConfigsResponse{}, nil
	}

	if req.InstallName != "" {
		return a.parseSingleInstall(installsPath, req.InstallName)
	}

	changedSet := make(map[string]bool, len(req.ChangedFiles))
	for _, f := range req.ChangedFiles {
		changedSet[f] = true
	}

	var result []ParsedInstallConfig

	entries, err := os.ReadDir(installsPath)
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
		installCfg, err := parseInstallTOML(fullPath)
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

func (a *Activities) parseSingleInstall(installsPath, installName string) (*ParseInstallConfigsResponse, error) {
	fileName := installName + ".toml"
	fullPath := filepath.Join(installsPath, fileName)

	if _, err := os.Stat(fullPath); err != nil {
		return &ParseInstallConfigsResponse{}, nil
	}

	installCfg, err := parseInstallTOML(fullPath)
	if err != nil {
		return nil, fmt.Errorf("unable to parse %s: %w", fileName, err)
	}

	return &ParseInstallConfigsResponse{
		Installs: []ParsedInstallConfig{
			{Config: installCfg, FilePath: fileName},
		},
	}, nil
}

func parseInstallTOML(filePath string) (*config.Install, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %w", filePath, err)
	}
	defer f.Close()

	cfg, err := configparse.ParseInstallConfig(f, validator.New())
	if err != nil {
		return nil, fmt.Errorf("unable to parse install config: %w", err)
	}

	return cfg, nil
}
