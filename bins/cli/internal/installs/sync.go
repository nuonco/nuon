package installs

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/pelletier/go-toml"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
)

func (s *Service) Sync(ctx context.Context, fileOrDir string, appID string) error {
	if fileOrDir == "" {
		return ui.PrintError(fmt.Errorf("file or directory path is required"))
	}

	installs, err := readInstallConfigs(fileOrDir)
	if err != nil {
		return ui.PrintError(err)
	}

	appID, err = lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	for _, install := range installs {
		appIns, err := s.syncInstall(ctx, install, appID)
		if err != nil {
			return ui.PrintError(fmt.Errorf("error syncing install %s: %w", install.Name, err))
		}

		ui.PrintSuccess(fmt.Sprintf("Install %s synced successfully", appIns.Name))
	}
	return nil
}

func (s *Service) syncInstall(ctx context.Context, install *config.Install, appID string) (*models.AppInstall, error) {
	if install == nil {
		return nil, fmt.Errorf("install cannot be nil")
	}

	isNew := false

	appInstall, err := s.api.GetInstall(ctx, install.Name)
	if err != nil {
		if !nuon.IsNotFound(err) {
			return nil, fmt.Errorf("error getting install %s: %w", install.Name, err)
		}
		isNew = true
	}

	if isNew {
		// Use defaults for any missing inputs.
		{
			appInputCfg, err := s.api.GetAppInputLatestConfig(ctx, appID)
			if err != nil {
				return nil, fmt.Errorf("error getting latest input config for app %s: %w", appID, err)
			}

			for _, ic := range appInputCfg.Inputs {
				val, ok := install.Inputs[ic.Name]
				if ok && val != "" {
					continue
				}
				if ic.Default != "" {
					install.Inputs[ic.Name] = ic.Default
				}
			}
		}

		req := models.ServiceCreateInstallRequest{
			Name:   &install.Name,
			Inputs: install.Inputs,
		}
		if install.AWSAccount != nil {
			req.AwsAccount = &models.ServiceCreateInstallRequestAwsAccount{
				Region: install.AWSAccount.Region,
			}
		}
		if install.ApprovalOption != config.InstallApprovalOptionUnknown {
			req.InstallConfig = &models.HelpersCreateInstallConfigParams{
				ApprovalOption: install.ApprovalOption.APIType(),
			}
		}
		appInstall, err = s.api.CreateInstall(ctx, appID, &req)
		if err != nil {
			return nil, fmt.Errorf("error creating install %s: %w", install.Name, err)
		}
	} else {
		if appInstall.AppID != appID {
			return nil, fmt.Errorf("install %s exists in a different app, aborting sync", install.Name)
		}

		if install.ApprovalOption != config.InstallApprovalOptionUnknown {
			if appInstall.InstallConfig == nil {
				appInstall.InstallConfig, err = s.api.CreateInstallConfig(ctx, appInstall.ID, &models.ServiceCreateInstallConfigRequest{
					ApprovalOption: install.ApprovalOption.APIType(),
				})
				if err != nil {
					return nil, err
				}
			} else {
				if appInstall.InstallConfig.ApprovalOption != install.ApprovalOption.APIType() {
					// Update the install config if the approval option has changed.
					_, err := s.api.UpdateInstallConfig(ctx, appInstall.ID, appInstall.InstallConfig.ID, &models.ServiceUpdateInstallConfigRequest{
						ApprovalOption: install.ApprovalOption.APIType(),
					})
					if err != nil {
						return nil, err
					}
				}
			}
		}

		currInputs, err := s.api.GetInstallCurrentInputs(ctx, appInstall.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting current inputs for install %s: %w", appInstall.Name, err)
		}
		// Use the current inputs as defaults, for missing values in the current inputs.
		for k, v := range currInputs.Values {
			if _, ok := install.Inputs[k]; !ok {
				install.Inputs[k] = v
			}
		}

		hasChanged := false
		if len(install.Inputs) != len(currInputs.Values) {
			hasChanged = true
		} else {
			// length is same, go through each input to see if any have changed.
			for k, v := range install.Inputs {
				if currInputs.Values[k] != v {
					hasChanged = true
					break
				}
			}
		}

		// If inputs have divereged, update the install inputs.
		if hasChanged {
			_, err = s.api.UpdateInstallInputs(ctx, appInstall.ID, &models.ServiceUpdateInstallInputsRequest{
				Inputs: install.Inputs,
			})
			if err != nil {
				return nil, fmt.Errorf("error updating inputs for install %s: %w", appInstall.Name, err)
			}
		}
	}

	return appInstall, nil
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
