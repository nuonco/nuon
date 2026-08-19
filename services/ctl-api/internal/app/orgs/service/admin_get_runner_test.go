package service

import (
	"context"
	"net/http"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *AdminGetOrgTestSuite) TestAdminGetOrgRunnerWithoutRunner() {
	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)
	org := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "org-without-runner",
		SandboxMode: true,
	}
	s.Require().NoError(s.service.DB.WithContext(ctx).Create(org).Error)
	s.Require().NoError(s.service.DB.WithContext(ctx).Create(&app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OwnerID:   org.ID,
		OwnerType: "orgs",
	}).Error)

	rr := s.makeRequest(http.MethodGet, "/v1/orgs/"+org.ID+"/admin-get-runner")

	s.Require().Equal(http.StatusNotFound, rr.Code, rr.Body.String())
}
