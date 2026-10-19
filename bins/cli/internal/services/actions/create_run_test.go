package actions

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type createRunAPI struct {
	nuon.Client
	roles          []*models.ServiceAvailableRole
	request        *models.ServiceCreateInstallActionWorkflowRunRequest
	installArg     string
	appArg         string
	actionArg      string
	configActionID string
	createdInstall string
	actionMissing  bool
	emptyPolls     int
	polls          int
	workflowStatus models.AppStatus
	runStatuses    []string
	runIndex       int
	logStreamID    string
	tailCalls      int
	recentInstall  string
	recentAction   string
	getRunInstall  string
}

func (a *createRunAPI) GetInstall(_ context.Context, id string) (*models.AppInstall, error) {
	a.installArg = id
	return &models.AppInstall{ID: "inst_resolved", AppID: "app_1"}, nil
}

func (a *createRunAPI) GetAppActionWorkflow(_ context.Context, appID, idOrName string) (*models.AppActionWorkflow, error) {
	a.appArg = appID
	a.actionArg = idOrName
	if a.actionMissing {
		return nil, notFoundError{}
	}
	return &models.AppActionWorkflow{ID: "act_resolved"}, nil
}

func (a *createRunAPI) GetAvailableRoles(_ context.Context, _ string) ([]*models.ServiceAvailableRole, error) {
	return a.roles, nil
}

func (a *createRunAPI) GetActionWorkflowLatestConfig(_ context.Context, id string) (*models.AppActionWorkflowConfig, error) {
	a.configActionID = id
	return &models.AppActionWorkflowConfig{ID: "awc_123"}, nil
}

func (a *createRunAPI) CreateInstallActionWorkflowRunWithResponse(_ context.Context, installID string, req *models.ServiceCreateInstallActionWorkflowRunRequest) (*models.AppWorkflowResponse, error) {
	a.createdInstall = installID
	a.request = req
	return &models.AppWorkflowResponse{WorkflowID: "wkf_1"}, nil
}

func (a *createRunAPI) GetInstallActionWorkflowRecentRuns(_ context.Context, installID, actionID string, _ *models.GetPaginatedQuery) (*models.AppInstallActionWorkflow, bool, error) {
	a.recentInstall = installID
	a.recentAction = actionID
	a.polls++
	if a.polls <= a.emptyPolls {
		return &models.AppInstallActionWorkflow{}, false, nil
	}
	return &models.AppInstallActionWorkflow{Runs: []*models.AppInstallActionWorkflowRun{
		{ID: "run_other", InstallWorkflowID: "wkf_other"},
		{ID: "run_1", InstallWorkflowID: "wkf_1"},
	}}, false, nil
}

func (a *createRunAPI) GetWorkflow(_ context.Context, _ string) (*models.AppWorkflow, error) {
	return &models.AppWorkflow{Status: &models.AppCompositeStatus{Status: a.workflowStatus}}, nil
}

func (a *createRunAPI) GetInstallActionWorkflowRun(_ context.Context, installID, _ string) (*models.AppInstallActionWorkflowRun, error) {
	a.getRunInstall = installID
	status := a.runStatuses[a.runIndex]
	if a.runIndex < len(a.runStatuses)-1 {
		a.runIndex++
	}
	run := &models.AppInstallActionWorkflowRun{ID: "run_1", Status: status}
	if a.logStreamID != "" {
		run.RunnerJob = &models.AppRunnerJob{LogStreamID: a.logStreamID}
	}
	return run, nil
}

func (a *createRunAPI) LogStreamTailLogs(_ context.Context, _ string, _ string, _ string, _ *nuon.LogStreamLogFilters) (*models.ServiceLogStreamTailLogsResponse, error) {
	a.tailCalls++
	if a.tailCalls > 1 {
		return &models.ServiceLogStreamTailLogsResponse{}, nil
	}
	return &models.ServiceLogStreamTailLogsResponse{Logs: []*models.AppOtelLogRecord{
		{Body: "hello", LogAttributes: map[string]string{"nuon.command_output": "true"}},
		{Body: "noise", LogAttributes: map[string]string{}},
	}}, nil
}

func (a *createRunAPI) GetLogStream(_ context.Context, _ string) (*models.AppLogStream, error) {
	return &models.AppLogStream{Open: false}, nil
}

type notFoundError struct{}

func (notFoundError) Error() string                         { return "not found" }
func (notFoundError) IsCode(code int) bool                  { return code == 404 }
func (notFoundError) IsServerError() bool                   { return false }
func (notFoundError) GetPayload() *models.StderrErrResponse { return &models.StderrErrResponse{} }

func newCreateRunService(api nuon.Client, installID string) *Service {
	v := viper.New()
	if installID != "" {
		v.Set("install_id", installID)
	}
	return New(validator.New(), api, &config.Config{Viper: v})
}

func fastPolls(t *testing.T) {
	prev := runPollInterval
	runPollInterval = time.Millisecond
	t.Cleanup(func() { runPollInterval = prev })
}

func TestCreateRunRejectsUnknownRole(t *testing.T) {
	api := &createRunAPI{roles: []*models.ServiceAvailableRole{
		{Name: "install-provision"},
		{Name: "install-maintenance"},
	}}
	service := newCreateRunService(api, "")

	err := service.CreateRun(context.Background(), CreateRunParams{InstallID: "inst_123", ActionWorkflowID: "action_123", Role: "provision"}, false)

	require.EqualError(t, err, `role "provision" is not available; available roles: install-maintenance, install-provision`)
	require.Nil(t, api.request)
}

func TestCreateRunForwardsRequestID(t *testing.T) {
	api := &createRunAPI{roles: []*models.ServiceAvailableRole{{Name: "install-maintenance"}}}
	service := newCreateRunService(api, "")
	params := CreateRunParams{InstallID: "inst_123", ActionWorkflowID: "action_123", Role: "install-maintenance", RequestID: "retry-1"}

	require.NoError(t, service.CreateRun(context.Background(), params, false))
	require.Equal(t, "retry-1", api.request.RequestID)

	params.RequestID = ""
	require.NoError(t, service.CreateRun(context.Background(), params, false))
	require.Empty(t, api.request.RequestID)
}

func TestCreateRunResolvesNames(t *testing.T) {
	api := &createRunAPI{}
	service := newCreateRunService(api, "")

	require.NoError(t, service.CreateRun(context.Background(), CreateRunParams{InstallID: "customer1", ActionWorkflowID: "pg_status"}, false))

	require.Equal(t, "customer1", api.installArg)
	require.Equal(t, "app_1", api.appArg)
	require.Equal(t, "pg_status", api.actionArg)
	require.Equal(t, "act_resolved", api.configActionID)
	require.Equal(t, "inst_resolved", api.createdInstall)
	require.Equal(t, "awc_123", *api.request.ActionWorkflowConfigID)
}

func TestCreateRunUsesSelectedInstall(t *testing.T) {
	api := &createRunAPI{}
	service := newCreateRunService(api, "inst_selected")

	require.NoError(t, service.CreateRun(context.Background(), CreateRunParams{ActionWorkflowID: "pg_status"}, false))
	require.Equal(t, "inst_selected", api.installArg)
}

func TestCreateRunRequiresInstall(t *testing.T) {
	api := &createRunAPI{}
	service := newCreateRunService(api, "")

	err := service.CreateRun(context.Background(), CreateRunParams{ActionWorkflowID: "pg_status"}, false)
	var userErr *ui.CLIUserError
	require.ErrorAs(t, err, &userErr)
	require.Nil(t, api.request)
}

func TestCreateRunMergesEnvFileAndFlags(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("A=file\nSHARED=file\n"), 0o600))
	api := &createRunAPI{}
	service := newCreateRunService(api, "inst_1")

	err := service.CreateRun(context.Background(), CreateRunParams{
		ActionWorkflowID: "pg_status",
		EnvFile:          envPath,
		Env:              []string{"SHARED=flag", "FOO=bar"},
	}, false)

	require.NoError(t, err)
	require.Equal(t, map[string]string{"A": "file", "SHARED": "flag", "FOO": "bar"}, api.request.RunEnvVars)
}

func TestCreateRunUnknownActionIsUserError(t *testing.T) {
	api := &createRunAPI{actionMissing: true}
	service := newCreateRunService(api, "inst_1")

	err := service.CreateRun(context.Background(), CreateRunParams{ActionWorkflowID: "nope"}, false)

	var userErr *ui.CLIUserError
	require.ErrorAs(t, err, &userErr)
	require.Equal(t, `action "nope" not found in app app_1`, userErr.Msg)
	require.Nil(t, api.request)
}

func TestCreateRunWaitFindsRunByWorkflowID(t *testing.T) {
	fastPolls(t)
	api := &createRunAPI{
		emptyPolls:     2,
		workflowStatus: models.AppStatusInDashProgress,
		runStatuses:    []string{"in-progress", "finished"},
		logStreamID:    "ls_1",
	}
	service := newCreateRunService(api, "inst_1")

	err := service.CreateRun(context.Background(), CreateRunParams{ActionWorkflowID: "pg_status", Wait: true}, true)

	require.NoError(t, err)
	require.Equal(t, 3, api.polls)
	require.Equal(t, "inst_resolved", api.recentInstall)
	require.Equal(t, "act_resolved", api.recentAction)
	require.Equal(t, "inst_resolved", api.getRunInstall)
	require.Equal(t, 5, api.tailCalls) // one data page, then drain polls
}

func TestCreateRunWaitStreamsOnlyCommandOutput(t *testing.T) {
	fastPolls(t)
	api := &createRunAPI{runStatuses: []string{"finished"}, logStreamID: "ls_1"}
	service := newCreateRunService(api, "inst_1")

	var out bytes.Buffer
	run, err := service.waitForRun(context.Background(), "inst_resolved", "run_1", &out, quietSpinner())
	require.NoError(t, err)
	require.Equal(t, "finished", run.Status)
	require.Equal(t, "hello\n", out.String())
}

func TestCreateRunWaitFailsWhenRunFails(t *testing.T) {
	fastPolls(t)
	api := &createRunAPI{runStatuses: []string{"error"}}
	service := newCreateRunService(api, "inst_1")

	err := service.CreateRun(context.Background(), CreateRunParams{ActionWorkflowID: "pg_status", Wait: true}, false)

	var userErr *ui.CLIUserError
	require.ErrorAs(t, err, &userErr)
	require.Equal(t, "action run run_1 ended with status error", userErr.Msg)
}

func TestCreateRunWaitErrorsWhenWorkflowEndsWithoutRun(t *testing.T) {
	fastPolls(t)
	api := &createRunAPI{emptyPolls: 100, workflowStatus: models.AppStatusError}
	service := newCreateRunService(api, "inst_1")

	err := service.CreateRun(context.Background(), CreateRunParams{ActionWorkflowID: "pg_status", Wait: true}, false)

	var userErr *ui.CLIUserError
	require.ErrorAs(t, err, &userErr)
	require.Equal(t, "action workflow wkf_1 ended with status error before a run was created", userErr.Msg)
}

func TestCreateRunWaitHonorsCancellation(t *testing.T) {
	fastPolls(t)
	api := &createRunAPI{emptyPolls: 1 << 30, workflowStatus: models.AppStatusInDashProgress}
	service := newCreateRunService(api, "inst_1")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := service.CreateRun(ctx, CreateRunParams{ActionWorkflowID: "pg_status", Wait: true}, false)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestGetRecentRunsResolvesNames(t *testing.T) {
	api := &createRunAPI{}
	service := newCreateRunService(api, "")

	require.NoError(t, service.GetRecentRuns(context.Background(), "customer1", "pg_status", 0, 20, true))
	require.Equal(t, "customer1", api.installArg)
	require.Equal(t, "app_1", api.appArg)
	require.Equal(t, "inst_resolved", api.recentInstall)
	require.Equal(t, "act_resolved", api.recentAction)
}

func TestGetRunResolvesInstallName(t *testing.T) {
	api := &createRunAPI{runStatuses: []string{"finished"}}
	service := newCreateRunService(api, "inst_selected")

	require.NoError(t, service.GetRun(context.Background(), "", "run_1", true))
	require.NoError(t, service.GetRun(context.Background(), "customer1", "run_1", true))
	require.Equal(t, "customer1", api.installArg)
	require.Equal(t, "inst_resolved", api.getRunInstall)
}
