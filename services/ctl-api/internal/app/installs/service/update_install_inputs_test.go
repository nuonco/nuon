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

// TestUpdateInstallInputsPartialMerge verifies the endpoint behaves as a true PATCH:
// a caller can send a subset of inputs, and the endpoint merges it with the install's
// existing inputs. Previously this failed because the request had to carry every required
// input.
func (s *InstallsServiceTestSuite) TestUpdateInstallInputsPartialMerge() {
	install := s.createTestInstall()

	var inputCfg app.AppInputConfig
	require.NoError(s.T(), s.deps.DB.
		Preload("AppInputs").
		Where("app_id = ?", s.testApp.ID).
		Order("created_at DESC").
		First(&inputCfg).Error)
	require.NotEmpty(s.T(), inputCfg.AppInputs)
	groupID := inputCfg.AppInputs[0].AppInputGroupID

	// The seeded config has a single (non-required) "region" input. Add a second
	// required vendor input and a customer (install_stack) sourced input so we can
	// exercise a partial update that omits a required input and preserves the
	// install_stack value.
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

	// Seed existing inputs covering all three, including the install_stack value.
	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), install.ID, inputCfg.ID, map[string]*string{
		"region": strPtr("us-west-2"),
		"size":   strPtr("large"),
		"vpc_id": strPtr("vpc-123"),
	})

	// Patch ONLY region — "size" is required but omitted, which previously failed.
	body := UpdateInstallInputsRequest{
		Inputs: map[string]*string{
			"region": strPtr("us-east-1"),
		},
	}
	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodPatch, path, body)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	// The new record reflects the merge: region updated, size + vpc_id preserved.
	var dbInputs app.InstallInputs
	require.NoError(s.T(), s.deps.DB.
		Where("install_id = ?", install.ID).
		Order("created_at DESC").
		First(&dbInputs).Error)
	assert.Equal(s.T(), "us-east-1", *dbInputs.Values["region"])
	assert.Equal(s.T(), "large", *dbInputs.Values["size"])
	assert.Equal(s.T(), "vpc-123", *dbInputs.Values["vpc_id"])
}

// TestUpdateInstallInputsRejectsInstallStackInput verifies that supplying a value for an
// install_stack (customer) sourced input is rejected, even when it is part of an otherwise
// valid partial update.
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
