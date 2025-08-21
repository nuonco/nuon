package installs

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
)

func (s *Service) Sync(ctx context.Context, fileOrDir string, appID string) error {
	if fileOrDir == "" {
		return ui.PrintError(fmt.Errorf("file or directory path is required"))
	}

	installCfgs, err := readInstallConfigs(fileOrDir)
	if err != nil {
		return ui.PrintError(err)
	}

	appID, err = lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	is := newAppInstallSyncer(s.api, appID)

	for _, installCfg := range installCfgs {
		sv := ui.NewSpinnerView(false)
		sv.Start(fmt.Sprintf("syncing install %s", installCfg.Name))

		appIns, err := is.syncInstall(ctx, installCfg, sv)
		if err != nil {
			err = fmt.Errorf("error syncing install %s: %w", installCfg.Name, err)
			sv.Fail(err)
			return err
		}
		sv.Success(fmt.Sprintf("install %s synced successfully", appIns.Name))
	}
	return nil
}

func readInstallConfigs(fileOrDir string) ([]*config.Install, error) {
	fileInfo, err := os.Stat(fileOrDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path '%s' does not exist.\n", fileOrDir)
		} else {
			return nil, fmt.Errorf("error accessing path '%s': %v\n", fileOrDir, err)
		}
	}

	if fileInfo.IsDir() {
		installs, err := readInstallConfigsFromDir(fileOrDir)
		if err != nil {
			return nil, err
		}

		return installs, nil
	} else if fileInfo.Mode().IsRegular() {
		install, err := parseInstallConfigFromFile(fileOrDir)
		if err != nil {
			return nil, fmt.Errorf("error reading install config from file '%s': %v", fileOrDir, err)
		}

		return []*config.Install{install}, nil
	}

	return nil, fmt.Errorf("Path '%s' is neither a regular file nor a directory (e.g., a symbolic link, device file).\n", fileOrDir)
}

func readInstallConfigsFromDir(fileOrDir string) ([]*config.Install, error) {
	installConfigs := make([]*config.Install, 0)

	err := filepath.Walk(fileOrDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path '%s': %v", path, err)
		}

		if info.IsDir() {
			return nil // Skip directories
		}

		if strings.HasSuffix(info.Name(), ".toml") {
			installConfig, err := parseInstallConfigFromFile(path)
			if err != nil {
				return fmt.Errorf("error reading install config from file '%s': %v", path, err)
			}
			installConfigs = append(installConfigs, installConfig)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing installs from directory '%s': %v", fileOrDir, err)
	}

	return installConfigs, nil
}

func parseInstallConfigFromFile(filePath string) (*config.Install, error) {
	byts, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file '%s': %w", filePath, err)
	}

	buf := bytes.NewReader(byts)
	tomlDec := toml.NewDecoder(buf)
	tomlDec.SetTagName("mapstructure")

	obj := make(map[string]interface{})
	err = tomlDec.Decode(&obj)
	if err != nil {
		return nil, fmt.Errorf("error decoding TOML from file '%s': %w", filePath, err)
	}

	// go from map[string]interface{} => config.Install
	var cfg config.Install
	mapDecCfg := config.DecoderConfig()
	mapDecCfg.Result = &cfg
	mapDec, err := mapstructure.NewDecoder(mapDecCfg)
	if err != nil {
		return nil, err
	}

	err = mapDec.Decode(obj)
	if err != nil {
		return nil, fmt.Errorf("error decoding config from file '%s': %w", filePath, err)
	}

	err = cfg.Parse()
	if err != nil {
		return nil, fmt.Errorf("error parsing config from file '%s': %w", filePath, err)
	}

	return &cfg, nil
}
