package helpers

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

func (s *Helpers) validateApp(parentApp *app.App) error {
	if len(parentApp.AppConfigs) < 1 {
		return stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any configs"),
			Description: "please make create at least one app config first",
		}
	}

	// validate the app is correctly configured and healthy
	if len(parentApp.AppSandboxConfigs) < 1 {
		return stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any sandbox configs"),
			Description: "please make create at least one app sandbox config first",
		}
	}
	if len(parentApp.AppRunnerConfigs) < 1 {
		return stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any runner configs"),
			Description: "please make create at least one app runner config first",
		}
	}

	if parentApp.Status == "error" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("app is in an error state"),
			Description: "can not create an install when app is in error state",
		}
	}

	return nil
}
