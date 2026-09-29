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
	install := s.createTestInstallWithActiveRunner()

	var inputCfg app.AppInputConfig
	require.NoError(s.T(), s.deps.DB.
		Where("app_id = ?", s.testApp.ID).
		Order("created_at DESC").
		First(&inputCfg).Error)

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

	require.NotNil(s.T(), inputs.WorkflowID)
	workflowID := *inputs.WorkflowID
	assert.NotEmpty(s.T(), workflowID)

	var dbInputs app.InstallInputs
	require.NoError(s.T(), s.deps.DB.
		Where("install_id = ?", install.ID).
		Order("created_at DESC").
		First(&dbInputs).Error)
	assert.Equal(s.T(), "us-east-1", *dbInputs.Values["region"])

	var dbWorkflow app.Workflow
	require.NoError(s.T(), s.deps.DB.Where("id = ?", workflowID).First(&dbWorkflow).Error)
	assert.Equal(s.T(), install.ID, dbWorkflow.OwnerID)
}

func (s *InstallsServiceTestSuite) TestUpdateInstallInputsPartialMerge() {
	install := s.createTestInstallWithActiveRunner()

	var inputCfg app.AppInputConfig
	require.NoError(s.T(), s.deps.DB.
		Preload("AppInputs").
		Where("app_id = ?", s.testApp.ID).
		Order("created_at DESC").
		First(&inputCfg).Error)
	require.NotEmpty(s.T(), inputCfg.AppInputs)
	groupID := inputCfg.AppInputs[0].AppInputGroupID

	requiredVendor := &app.AppInput{
		AppInputConfigID: inputCfg.ID,
		AppInputGroupID:  groupID,
		Name:             "size",
		DisplayName:      "Size",
		Type:             app.AppInputTypeString,
		Required:         true,
		Source:           app.AppInputSourceVendor,
	}
	require.NoError(s.T(), s.deps.DB.Create(requiredVendor).Error)
	customerInput := &app.AppInput{
		AppInputConfigID: inputCfg.ID,
		AppInputGroupID:  groupID,
		Name:             "vpc_id",
		DisplayName:      "VPC ID",
		Type:             app.AppInputTypeString,
		Source:           app.AppInputSourceCustomer,
	}
	require.NoError(s.T(), s.deps.DB.Create(customerInput).Error)

	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), install.ID, inputCfg.ID, map[string]*string{
		"region": strPtr("us-west-2"),
		"size":   strPtr("large"),
		"vpc_id": strPtr("vpc-123"),
	})

	body := UpdateInstallInputsRequest{
		Inputs: map[string]*string{
			"region": strPtr("us-east-1"),
		},
	}
	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodPatch, path, body)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var dbInputs app.InstallInputs
	require.NoError(s.T(), s.deps.DB.
		Where("install_id = ?", install.ID).
		Order("created_at DESC").
		First(&dbInputs).Error)
	assert.Equal(s.T(), "us-east-1", *dbInputs.Values["region"])
	assert.Equal(s.T(), "large", *dbInputs.Values["size"])
	assert.Equal(s.T(), "vpc-123", *dbInputs.Values["vpc_id"])
}

func (s *InstallsServiceTestSuite) TestUpdateInstallInputsRejectsInstallStackInput() {
	install := s.createTestInstall()

	var inputCfg app.AppInputConfig
	require.NoError(s.T(), s.deps.DB.
		Preload("AppInputs").
		Where("app_id = ?", s.testApp.ID).
		Order("created_at DESC").
		First(&inputCfg).Error)
	require.NotEmpty(s.T(), inputCfg.AppInputs)
	groupID := inputCfg.AppInputs[0].AppInputGroupID

	customerInput := &app.AppInput{
		AppInputConfigID: inputCfg.ID,
		AppInputGroupID:  groupID,
		Name:             "vpc_id",
		DisplayName:      "VPC ID",
		Type:             app.AppInputTypeString,
		Source:           app.AppInputSourceCustomer,
	}
	require.NoError(s.T(), s.deps.DB.Create(customerInput).Error)

	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), install.ID, inputCfg.ID, map[string]*string{
		"region": strPtr("us-west-2"),
	})

	body := UpdateInstallInputsRequest{
		Inputs: map[string]*string{
			"vpc_id": strPtr("vpc-999"),
		},
	}
	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodPatch, path, body)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code, "body: %s", rr.Body.String())
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
