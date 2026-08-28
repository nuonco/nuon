package installs

import (
	"context"
	"encoding/json"
	"os"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) Register(ctx context.Context, path string, asJSON bool) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return ui.PrintError(err)
	}
	var registration models.CustomermanagedInstallationRegistration
	if err := json.Unmarshal(contents, &registration); err != nil {
		return ui.PrintError(err)
	}
	result, err := s.api.RegisterInstall(ctx, &registration)
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	ui.NewGetView().Render([][]string{
		{"install id", result.Install.ID},
		{"install name", result.Install.Name},
		{"registration id", result.Registration.ID},
		{"release id", result.ReleaseDeployment.ReleaseID},
		{"package id", result.ReleaseDeployment.PackageID},
		{"connectivity", result.ManagementPolicy.Connectivity},
	})
	return nil
}
