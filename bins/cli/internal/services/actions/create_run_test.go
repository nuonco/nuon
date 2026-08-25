package actions

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type createRunAPI struct {
	nuon.Client
	roles   []*models.ServiceAvailableRole
	request *models.ServiceCreateInstallActionWorkflowRunRequest
}

func (a *createRunAPI) GetAvailableRoles(_ context.Context, _ string) ([]*models.ServiceAvailableRole, error) {
	return a.roles, nil
}

func (a *createRunAPI) GetActionWorkflowLatestConfig(_ context.Context, _ string) (*models.AppActionWorkflowConfig, error) {
	return &models.AppActionWorkflowConfig{ID: "awc_123"}, nil
}

func (a *createRunAPI) CreateInstallActionWorkflowRun(_ context.Context, _ string, req *models.ServiceCreateInstallActionWorkflowRunRequest) error {
	a.request = req
	return nil
}

func TestCreateRunRejectsUnknownRole(t *testing.T) {
	api := &createRunAPI{roles: []*models.ServiceAvailableRole{
		{Name: "install-provision"},
		{Name: "install-maintenance"},
	}}
	service := New(validator.New(), api, &config.Config{Viper: viper.New()})

	err := service.CreateRun(context.Background(), "inst_123", "action_123", "provision", false)

	require.EqualError(t, err, `role "provision" is not available; available roles: install-maintenance, install-provision`)
	require.Nil(t, api.request)
}
