package service

import (
	"fmt"
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *ComponentsServiceTestSuite) TestCancelAppComponentBuildWithoutQueueSignal() {
	cmp := s.getSeededComponent(app.ComponentTypeHelmChart)
	conn := s.getSeededConfigConnection(cmp.ID)
	build := s.deps.Seeder.CreateComponentBuild(s.ctx, s.T(), conn.ID)

	path := fmt.Sprintf("/v1/apps/%s/components/%s/builds/%s/cancel", s.testApp.ID, cmp.ID, build.ID)
	rr := s.makeRequest(http.MethodPost, path, nil)

	s.Require().Equal(http.StatusConflict, rr.Code, rr.Body.String())
}
