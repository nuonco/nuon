package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestUpdateInstallInputsSuccess() {
	install := s.createTestInstall()

	// Find the AppInputConfig created by CreateAppConfig.
	var inputCfg app.AppInputConfig
	require.NoError(s.T(), s.deps.DB.
		Where("app_id = ?", s.testApp.ID).
		Order("created_at DESC").
		First(&inputCfg).Error)

	// Seed existing inputs so the update has something to merge with.
	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), install.ID, inputCfg.ID, map[string]*string{
		"region": strPtr("us-west-2"),
	})

	body := UpdateInstallInputsRequest{
		Inputs: map[string]*string{
			"region": strPtr("us-east-1"),
		},
	}

	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodPatch, path, body)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var inputs app.InstallInputs
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &inputs))
	assert.NotEmpty(s.T(), inputs.ID)

	// Workflow should have been created for the input update.
	require.NotNil(s.T(), inputs.WorkflowID)
	workflowID := *inputs.WorkflowID
	assert.NotEmpty(s.T(), workflowID)

	// Verify the new inputs record was persisted.
	var dbInputs app.InstallInputs
	require.NoError(s.T(), s.deps.DB.
		Where("install_id = ?", install.ID).
		Order("created_at DESC").
		First(&dbInputs).Error)
	assert.Equal(s.T(), "us-east-1", *dbInputs.Values["region"])

	// Verify the workflow was created.
	var dbWorkflow app.Workflow
	require.NoError(s.T(), s.deps.DB.Where("id = ?", workflowID).First(&dbWorkflow).Error)
	assert.Equal(s.T(), install.ID, dbWorkflow.OwnerID)
}

func (s *InstallsServiceTestSuite) TestUpdateInstallInputsNoExistingInputs() {
	install := s.createTestInstall()

	body := UpdateInstallInputsRequest{
		Inputs: map[string]*string{
			"region": strPtr("us-east-1"),
		},
	}

	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodPatch, path, body)
	// Should fail because there are no existing inputs to update.
	// The wrapped gorm.ErrRecordNotFound propagates through the error middleware as 404.
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestUpdateInstallInputsNotFound() {
	body := UpdateInstallInputsRequest{
		Inputs: map[string]*string{
			"region": strPtr("us-east-1"),
		},
	}

	rr := s.makeRequest(http.MethodPatch, "/v1/installs/nonexistent/inputs", body)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}
