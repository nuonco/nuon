package converters

import (
	"github.com/powertoolsdev/mono/services/api/internal/models"
	appv1 "github.com/powertoolsdev/mono/pkg/protos/api/generated/types/app/v1"
)

// App model to proto converts app domain model into app proto message
func AppModelToProto(app *models.App) *appv1.App {
	return &appv1.App{
		Id:              app.ID.String(),
		Name:            app.Name,
		GithubInstallId: app.GithubInstallID,
		OrgId:           app.OrgID.String(),
		CreatedById:     app.CreatedByID,
		UpdatedAt:       TimeToDatetime(app.UpdatedAt),
		CreatedAt:       TimeToDatetime(app.CreatedAt),
	}
}

// AppModelsToProtos converts a slice of app models to protos
func AppModelsToProtos(apps []*models.App) []*appv1.App {
	protos := make([]*appv1.App, len(apps))
	for idx, app := range apps {
		protos[idx] = AppModelToProto(app)
	}

	return protos
}
