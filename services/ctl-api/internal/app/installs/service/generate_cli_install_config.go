package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	"github.com/pelletier/go-toml"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
)

// @ID						GenerateCLIInstallConfig
// @Summary				generate an install config to be used with CLI
// @Description.markdown	generate_cli_install_config.md
// @Param					install_id		path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				application/octet-stream
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{file}	config.Install
// @Router					/v1/installs/{install_id}/generate-cli-install-config [get]
func (s *service) GenerateCLIInstallConfig(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	installCfg, err := s.genCLIInstallConfig(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("error generating config from current state: %w", err))
		return
	}

	var response bytes.Buffer
	enc := toml.NewEncoder(&response)
	enc.SetTagName("mapstructure")
	enc.Order(toml.OrderAlphabetical)

	err = enc.Encode(installCfg)
	if err != nil {
		ctx.Error(fmt.Errorf("error encoding config: %w", err))
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.toml\"", strcase.ToSnake(installCfg.Name)))
	ctx.Data(http.StatusOK, "application/octet-stream", response.Bytes())
}

func (s *service) genCLIInstallConfig(ctx context.Context, installID string) (*config.Install, error) {
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install %s: %w", installID, err)
	}

	installCfg := config.Install{
		Name: install.Name,
	}

	if install.AWSAccount != nil {
		installCfg.AWSAccount = &config.AWSAccount{
			Region: install.AWSAccount.Region,
		}
	}

	installConfig, err := s.helpers.GetLatestInstallConfig(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("failed parsing approval option: %w", err)
	}

	if installConfig != nil {
		installCfg.ApprovalOption = config.InstallApprovalOption(installConfig.ApprovalOption)
	}

	appInputCfg, err := s.helpers.GetPinnedAppInputConfig(ctx, install.AppID, install.AppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to get app input config for install %s: %w", installID, err)
	}

	installInputs, err := s.getLatestInstallInputs(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get inputs for install %s: %w", installID, err)
	}

	inputGroups := make(map[string]config.InputGroup)
	for _, inp := range appInputCfg.AppInputs {
		if inputGroups[inp.AppInputGroup.Name] == nil {
			inputGroups[inp.AppInputGroup.Name] = make(config.InputGroup)
		}
		if inp.Sensitive {
			continue
		}

		val, ok := installInputs.Values[inp.Name]
		if !ok {
			s.l.Error("input is not set when generating install config",
				zap.String("key", inp.Name),
			)

			if inp.Required {
				inputGroups[inp.AppInputGroup.Name][inp.Name] = ""
			}
		} else {
			inputGroups[inp.AppInputGroup.Name][inp.Name] = generics.FromPtrStr(val)
		}
	}

	for groupName, inputGroupInputs := range inputGroups {
		if len(inputGroupInputs) > 0 {
			// note(sk): we're doing this here to maintain backward compatibility,
			// currently group name is being dropped and is not being passed forward
			inputGroupInputs["__nuon.input.group"] = groupName
			installCfg.InputGroups = append(installCfg.InputGroups, inputGroupInputs)
		}
	}

	return &installCfg, nil
}
