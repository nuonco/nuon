package service

import "github.com/nuonco/nuon/services/ctl-api/internal/app"

func mcpAppWithDerivedStatus(a *app.App) *app.App {
	result := *a
	if a.StatusV2.Status != "" {
		result.Status = app.AppStatus(a.StatusV2.Status)
		result.StatusDescription = a.StatusV2.StatusHumanDescription
	}
	return &result
}
