package installs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/pelletier/go-toml/v2"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/config"
)

func (s *Service) Sync(ctx context.Context, fileOrDir string, appID string, confirm, approveAll, wait, dryRun, asJSON bool) error {
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

	curInstalls, err := s.listAllAppInstalls(ctx, appID)
	if err != nil {
		return ui.PrintError(fmt.Errorf("error listing installs for app %s: %w", appID, err))
	}

	is := newAppInstallSyncer(s.api, appID, s.cfg.OrgID, s.cfg.Interactive, asJSON, approveAll)

	results := make([]syncedInstall, 0, len(installCfgs))
	for _, installCfg := range installCfgs {
		var installID string

		appInstall, ok := curInstalls[installCfg.Name]
		if ok && appInstall != nil {
			installID = appInstall.ID

			if appInstall.AppID != appID {
				if !asJSON {
					ui.PrintWarning(fmt.Sprintf("install %s is not associated with app %s", installCfg.Name, appID))
				}
				results = append(results, syncedInstall{Name: installCfg.Name, InstallID: installID, Status: "skipped"})
				continue
			}

			// Check if the install is managed by Nuon CLI config.
			// If not, return an error to prevent overwriting.
			if appInstall.Metadata["managed_by"] != ManagedByNuonCLIConfig {
				return ui.PrintError(fmt.Errorf("install %s is not managed by an install config file, aborting sync to prevent overwriting", installCfg.Name))
			}
		}

		synced, err := is.syncInstall(ctx, installCfg, installID, confirm, wait, dryRun)
		if err != nil {
			return ui.PrintError(fmt.Errorf("error syncing install %s: %w", installCfg.Name, err))
		}

		results = append(results, syncResultFor(installCfg.Name, synced, dryRun))
	}

	if asJSON {
		ui.PrintJSON(syncResult{Installs: results})
	}
	return nil
}

type syncResult struct {
	Installs []syncedInstall `json:"installs"`
}

type syncedInstall struct {
	Name      string `json:"name"`
	InstallID string `json:"install_id,omitempty"`
	Status    string `json:"status"`
}

func syncResultFor(name string, synced *models.AppInstall, dryRun bool) syncedInstall {
	switch {
	case dryRun:
		return syncedInstall{Name: name, Status: "dry_run"}
	case synced == nil:
		return syncedInstall{Name: name, Status: "skipped"}
	default:
		return syncedInstall{Name: name, InstallID: synced.ID, Status: "synced"}
	}
}

func (s *Service) listAllAppInstalls(ctx context.Context, appID string) (map[string]*models.AppInstall, error) {
	var (
		hasMore = true
		offset  int
		limit   = 50
		result  = make(map[string]*models.AppInstall)
	)
	for hasMore {
		installs, more, err := s.api.GetAppInstalls(ctx, appID, &models.GetPaginatedQuery{
			Offset: offset,
			Limit:  limit,
		})
		if err != nil {
			return nil, err
		}
		for _, install := range installs {
			result[install.Name] = install
		}
		offset += limit
		hasMore = more
	}
	return result, nil
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
	cfg, err := parseInstallConfig(buf)
	if err != nil {
		return nil, fmt.Errorf("error parsing config from file '%s': %w", filePath, err)
	}

	return cfg, nil
}

func parseInstallConfig(raw io.Reader) (*config.Install, error) {
	tomlDec := toml.NewDecoder(raw)

	obj := make(map[string]interface{})
	err := tomlDec.Decode(&obj)
	if err != nil {
		return nil, fmt.Errorf("error decoding TOML: %w", err)
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
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	err = cfg.Parse()
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return &cfg, nil
}
