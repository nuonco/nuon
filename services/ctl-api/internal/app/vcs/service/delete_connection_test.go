package service

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *VCSServiceTestSuite) TestDeleteConnection_Success() {
	// Create a test connection
	conn := s.createTestVCSConnection()

	rr := s.makeRequest(http.MethodDelete, fmt.Sprintf("/v1/vcs/connections/%s", conn.ID), nil)

	if rr.Code != http.StatusNoContent {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNoContent, rr.Code)

	// Verify connection is deleted from database
	var deletedConn app.VCSConnection
	err := s.service.DB.Where("id = ?", conn.ID).First(&deletedConn).Error
	require.Error(s.T(), err)
	require.Equal(s.T(), gorm.ErrRecordNotFound, err)
}
