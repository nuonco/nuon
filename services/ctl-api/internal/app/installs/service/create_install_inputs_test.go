package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *InstallsServiceTestSuite) TestCreateInstallInputsSuccess() {
	install := s.createTestInstall()

	body := CreateInstallInputsRequest{
		Inputs: map[string]*string{
			"region": strPtr("us-east-1"),
		},
	}

	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	require.Equal(s.T(), http.StatusCreated, rr.Code, "body: %s", rr.Body.String())

	var inputs app.InstallInputs
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &inputs))
	assert.NotEmpty(s.T(), inputs.ID)
	assert.Equal(s.T(), install.ID, inputs.InstallID)

	// Verify record persisted in DB.
	var dbInputs app.InstallInputs
	require.NoError(s.T(), s.deps.DB.Where("id = ?", inputs.ID).First(&dbInputs).Error)
	assert.Equal(s.T(), install.ID, dbInputs.InstallID)
	assert.Equal(s.T(), "us-east-1", *dbInputs.Values["region"])
}

func (s *InstallsServiceTestSuite) TestCreateInstallInputsEmptyInputs() {
	install := s.createTestInstall()

	body := CreateInstallInputsRequest{
		Inputs: map[string]*string{},
	}

	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *InstallsServiceTestSuite) TestCreateInstallInputsNotFound() {
	body := CreateInstallInputsRequest{
		Inputs: map[string]*string{
			"region": strPtr("us-east-1"),
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs/nonexistent/inputs", body)
	assert.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestCreateInstallInputsInvalidKey() {
	install := s.createTestInstall()

	body := CreateInstallInputsRequest{
		Inputs: map[string]*string{
			"nonexistent_key": strPtr("value"),
		},
	}

	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	assert.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func strPtr(s string) *string {
	return &s
}
