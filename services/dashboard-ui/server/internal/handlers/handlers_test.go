package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/handlers"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// apiResponse mirrors the TAPIResponse shape.
type apiResponse struct {
	Data    json.RawMessage `json:"data"`
	Error   json.RawMessage `json:"error"`
	Status  int             `json:"status"`
	Headers json.RawMessage `json:"headers"`
}

// HandlersSuite is the test suite for BFF handlers.
type HandlersSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	mockClient *MockClient
	engine     *gin.Engine
}

func (s *HandlersSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockClient = NewMockClient(s.ctrl)
	s.engine = gin.New()
}

func (s *HandlersSuite) TearDownTest() {
	s.ctrl.Finish()
}

// injectClient returns middleware that sets mock client on context.
func (s *HandlersSuite) injectClient() gin.HandlerFunc {
	return func(c *gin.Context) {
		cctx.SetAPIClientGinContext(c, s.mockClient)
		c.Next()
	}
}

// doRequest performs an HTTP request against the test engine and returns the response.
func (s *HandlersSuite) doRequest(method, path string, body ...string) *httptest.ResponseRecorder {
	var req *http.Request
	if len(body) > 0 {
		req = httptest.NewRequest(method, path, strings.NewReader(body[0]))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)
	return w
}

// parseResponse parses the response body into the TAPIResponse shape.
func (s *HandlersSuite) parseResponse(w *httptest.ResponseRecorder) apiResponse {
	var resp apiResponse
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

// --- Health Handler Tests ---

func (s *HandlersSuite) TestHealthLivez() {
	cfg := &internal.Config{Version: "v1.0.0", GitRef: "abc123"}
	h := handlers.NewHealthHandler(cfg)
	s.Require().NoError(h.RegisterRoutes(s.engine))

	w := s.doRequest("GET", "/livez")
	s.Equal(http.StatusOK, w.Code)

	var body map[string]string
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &body))
	s.Equal("ok", body["status"])
}

func (s *HandlersSuite) TestHealthVersion() {
	cfg := &internal.Config{Version: "v1.0.0", GitRef: "abc123"}
	h := handlers.NewHealthHandler(cfg)
	s.Require().NoError(h.RegisterRoutes(s.engine))

	w := s.doRequest("GET", "/version")
	s.Equal(http.StatusOK, w.Code)

	var body map[string]string
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &body))
	s.Equal("v1.0.0", body["version"])
	s.Equal("abc123", body["git_ref"])
}

// --- Apps Handler Tests ---

func (s *HandlersSuite) TestGetApps() {
	l := zap.NewNop()
	h := handlers.NewAppsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	expected := []*models.AppApp{{ID: "app1", Name: "test-app"}}
	s.mockClient.EXPECT().SetOrgID("org1")
	s.mockClient.EXPECT().GetApps(gomock.Any(), gomock.Any()).Return(expected, false, nil)

	w := s.doRequest("GET", "/api/orgs/org1/apps")
	s.Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Equal(http.StatusOK, resp.Status)
	s.Equal("null", string(resp.Error))

	var apps []*models.AppApp
	s.Require().NoError(json.Unmarshal(resp.Data, &apps))
	s.Len(apps, 1)
	s.Equal("app1", apps[0].ID)
}

func (s *HandlersSuite) TestGetApp() {
	l := zap.NewNop()
	h := handlers.NewAppsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	expected := &models.AppApp{ID: "app1", Name: "test-app"}
	s.mockClient.EXPECT().SetOrgID("org1")
	s.mockClient.EXPECT().GetApp(gomock.Any(), "app1").Return(expected, nil)

	w := s.doRequest("GET", "/api/orgs/org1/apps/app1")
	s.Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Equal(http.StatusOK, resp.Status)
	s.Equal("null", string(resp.Error))
}

func (s *HandlersSuite) TestGetAppsError() {
	l := zap.NewNop()
	h := handlers.NewAppsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	s.mockClient.EXPECT().SetOrgID("org1")
	s.mockClient.EXPECT().GetApps(gomock.Any(), gomock.Any()).Return(nil, false, fmt.Errorf("api error"))

	w := s.doRequest("GET", "/api/orgs/org1/apps")
	s.Equal(http.StatusInternalServerError, w.Code)

	resp := s.parseResponse(w)
	s.Equal(http.StatusInternalServerError, resp.Status)
	s.NotEqual("null", string(resp.Error))
}

// --- Orgs Handler Tests ---

func (s *HandlersSuite) TestGetOrgs() {
	l := zap.NewNop()
	h := handlers.NewOrgsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	expected := []*models.AppOrg{{ID: "org1", Name: "test-org"}}
	s.mockClient.EXPECT().GetOrgs(gomock.Any(), gomock.Any()).Return(expected, false, nil)

	w := s.doRequest("GET", "/api/orgs")
	s.Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Equal(http.StatusOK, resp.Status)
	s.Equal("null", string(resp.Error))
}

func (s *HandlersSuite) TestGetOrgFeatures() {
	l := zap.NewNop()
	h := handlers.NewOrgsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	w := s.doRequest("GET", "/api/orgs/org1/features")
	s.Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Equal("[]", string(resp.Data))
}

// --- Account Handler Tests ---

func (s *HandlersSuite) TestGetAccount() {
	l := zap.NewNop()
	h := handlers.NewAccountHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	expected := &models.AppAccount{ID: "acct1"}
	s.mockClient.EXPECT().GetCurrentUser(gomock.Any()).Return(expected, nil)

	w := s.doRequest("GET", "/api/account")
	s.Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Equal(http.StatusOK, resp.Status)
}

// --- Installs Handler Tests ---

func (s *HandlersSuite) TestGetInstalls() {
	l := zap.NewNop()
	h := handlers.NewInstallsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	expected := []*models.AppInstall{{ID: "inst1"}}
	s.mockClient.EXPECT().SetOrgID("org1")
	s.mockClient.EXPECT().GetAllInstalls(gomock.Any(), gomock.Any()).Return(expected, false, nil)

	w := s.doRequest("GET", "/api/orgs/org1/installs")
	s.Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Equal(http.StatusOK, resp.Status)
}

// --- Actions Handler Tests ---

func (s *HandlersSuite) TestBuildComponent() {
	l := zap.NewNop()
	h := handlers.NewActionsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	expected := &models.AppComponentBuild{ID: "bld1"}
	s.mockClient.EXPECT().SetOrgID("org1")
	s.mockClient.EXPECT().CreateComponentBuild(gomock.Any(), "comp1", gomock.Any()).Return(expected, nil)

	body := `{"componentId":"comp1","orgId":"org1"}`
	w := s.doRequest("POST", "/api/actions/apps/build-component", body)
	s.Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Equal(http.StatusOK, resp.Status)
	s.Equal("null", string(resp.Error))
}

func (s *HandlersSuite) TestCreateOrg() {
	l := zap.NewNop()
	h := handlers.NewActionsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	expected := &models.AppOrg{ID: "org-new", Name: "my-org"}
	s.mockClient.EXPECT().CreateOrg(gomock.Any(), gomock.Any()).Return(expected, nil)

	body := `{"name":"my-org"}`
	w := s.doRequest("POST", "/api/actions/orgs/create-org", body)
	s.Equal(http.StatusOK, w.Code)

	resp := s.parseResponse(w)
	s.Equal(http.StatusOK, resp.Status)
}

func (s *HandlersSuite) TestBadRequestBody() {
	l := zap.NewNop()
	h := handlers.NewActionsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(h.RegisterRoutes(s.engine))

	w := s.doRequest("POST", "/api/actions/apps/build-component", "not-json")
	s.Equal(http.StatusBadRequest, w.Code)
}

// --- Response Format Tests ---

func (s *HandlersSuite) TestResponseFormatSuccess() {
	cfg := &internal.Config{Version: "v1", GitRef: "ref"}
	h := handlers.NewHealthHandler(cfg)
	// Health endpoints don't use TAPIResponse, test with apps instead
	l := zap.NewNop()
	ah := handlers.NewAppsHandler(l)
	s.engine.Use(s.injectClient())
	_ = h // unused, just checking we can construct it
	s.Require().NoError(ah.RegisterRoutes(s.engine))

	s.mockClient.EXPECT().SetOrgID("org1")
	s.mockClient.EXPECT().GetApps(gomock.Any(), gomock.Any()).Return([]*models.AppApp{}, false, nil)

	w := s.doRequest("GET", "/api/orgs/org1/apps")
	s.Equal(http.StatusOK, w.Code)

	// Verify TAPIResponse shape has all 4 keys
	var raw map[string]json.RawMessage
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &raw))
	s.Contains(raw, "data")
	s.Contains(raw, "error")
	s.Contains(raw, "status")
	s.Contains(raw, "headers")
}

func (s *HandlersSuite) TestResponseFormatError() {
	l := zap.NewNop()
	ah := handlers.NewAppsHandler(l)
	s.engine.Use(s.injectClient())
	s.Require().NoError(ah.RegisterRoutes(s.engine))

	s.mockClient.EXPECT().SetOrgID("org1")
	s.mockClient.EXPECT().GetApps(gomock.Any(), gomock.Any()).Return(nil, false, fmt.Errorf("something broke"))

	w := s.doRequest("GET", "/api/orgs/org1/apps")
	s.Equal(http.StatusInternalServerError, w.Code)

	var raw map[string]json.RawMessage
	s.Require().NoError(json.Unmarshal(w.Body.Bytes(), &raw))
	s.Contains(raw, "data")
	s.Contains(raw, "error")
	s.Contains(raw, "status")
	s.Contains(raw, "headers")

	// data should be null on error
	s.Equal("null", string(raw["data"]))

	// error should contain error and description
	var errBody map[string]string
	s.Require().NoError(json.Unmarshal(raw["error"], &errBody))
	s.Contains(errBody, "error")
	s.Contains(errBody, "description")
	s.Equal("something broke", errBody["error"])
}

// --- No Client on Context ---

func (s *HandlersSuite) TestNoClientReturnsError() {
	l := zap.NewNop()
	ah := handlers.NewAppsHandler(l)
	// Do NOT inject client middleware
	s.Require().NoError(ah.RegisterRoutes(s.engine))

	w := s.doRequest("GET", "/api/orgs/org1/apps")
	s.Equal(http.StatusInternalServerError, w.Code)
}

func TestHandlers(t *testing.T) {
	suite.Run(t, new(HandlersSuite))
}
