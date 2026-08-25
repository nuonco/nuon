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
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type adHocAPI struct {
	nuon.Client
	installID     string
	request       *models.ServiceCreateAdHocActionRequest
	roles         []*models.ServiceAvailableRole
	runs          []*models.AppInstallActionWorkflowRun
	runIndex      int
	logPages      []adHocLogPage
	logPageIndex  int
	logCursors    []string
	logStreamOpen bool
}

type adHocLogPage struct {
	records []*models.AppOtelLogRecord
	next    string
	hasMore bool
}

func (a *adHocAPI) GetInstall(_ context.Context, installID string) (*models.AppInstall, error) {
	a.installID = installID
	return &models.AppInstall{ID: "inst_resolved"}, nil
}

func (a *adHocAPI) GetAvailableRoles(_ context.Context, _ string) ([]*models.ServiceAvailableRole, error) {
	return a.roles, nil
}

func (a *adHocAPI) CreateAdHocAction(_ context.Context, installID string, req *models.ServiceCreateAdHocActionRequest) (*models.ServiceCreateAdHocActionResponse, error) {
	a.installID = installID
	a.request = req
	return &models.ServiceCreateAdHocActionResponse{
		ID:         "run_123",
		InstallID:  installID,
		WorkflowID: "workflow_123",
		Status:     models.AppInstallActionWorkflowRunStatusQueued,
	}, nil
}

func (a *adHocAPI) GetInstallActionWorkflowRun(_ context.Context, _, _ string) (*models.AppInstallActionWorkflowRun, error) {
	run := a.runs[a.runIndex]
	if a.runIndex < len(a.runs)-1 {
		a.runIndex++
	}
	return run, nil
}

func (a *adHocAPI) LogStreamTailLogs(_ context.Context, _ string, cursor string, _ string) (*models.ServiceLogStreamTailLogsResponse, error) {
	a.logCursors = append(a.logCursors, cursor)
	page := a.logPages[a.logPageIndex]
	if a.logPageIndex < len(a.logPages)-1 {
		a.logPageIndex++
	}
	return &models.ServiceLogStreamTailLogsResponse{
		Logs:    page.records,
		Next:    page.next,
		HasMore: page.hasMore,
	}, nil
}

func (a *adHocAPI) GetLogStream(_ context.Context, _ string) (*models.AppLogStream, error) {
	return &models.AppLogStream{Open: a.logStreamOpen}, nil
}

func TestCreateAdHocRunUsesSelectedInstallAndBuildsRequest(t *testing.T) {
	scriptPath := filepath.Join(t.TempDir(), "debug.sh")
	require.NoError(t, os.WriteFile(scriptPath, []byte("echo hello\n"), 0o600))

	v := viper.New()
	v.Set("install_id", "selected-install")
	cfg := &config.Config{Viper: v}
	api := &adHocAPI{roles: []*models.ServiceAvailableRole{{Name: "maintenance"}}}
	service := New(validator.New(), api, cfg)

	err := service.CreateAdHocRun(context.Background(), AdHocParams{
		ScriptPath:       scriptPath,
		Env:              []string{"DEBUG=true"},
		Timeout:          10 * time.Minute,
		Name:             "debug script",
		Role:             "maintenance",
		EnableKubeConfig: false,
	}, true)
	require.NoError(t, err)
	require.Equal(t, "inst_resolved", api.installID)
	require.Equal(t, "echo hello\n", api.request.InlineContents)
	require.Empty(t, api.request.Command)
	require.Equal(t, map[string]string{"DEBUG": "true"}, api.request.EnvVars)
	require.Equal(t, int64(600), api.request.Timeout)
	require.Equal(t, "debug script", api.request.Name)
	require.Equal(t, "maintenance", api.request.Role)
	require.NotNil(t, api.request.EnableKubeConfig)
	require.False(t, *api.request.EnableKubeConfig)
}

func TestCreateAdHocRunRejectsUnknownRole(t *testing.T) {
	api := &adHocAPI{roles: []*models.ServiceAvailableRole{
		{Name: "install-provision"},
		{Name: "install-maintenance"},
	}}
	service := New(validator.New(), api, &config.Config{Viper: viper.New()})

	err := service.CreateAdHocRun(context.Background(), AdHocParams{
		InstallID: "inst_123",
		Command:   "echo hello",
		Timeout:   time.Minute,
		Role:      "provision",
	}, false)

	require.EqualError(t, err, `role "provision" is not available; available roles: install-maintenance, install-provision`)
	require.Nil(t, api.request)
}

func TestParseAdHocEnvMergesFileAndFlags(t *testing.T) {
	envPath := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("# defaults\nFROM_FILE='value'\nexport SHARED=file\nEMPTY=\n"), 0o600))

	env, err := parseAdHocEnv(envPath, []string{"SHARED=flag", "WITH_EQUALS=one=two"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"FROM_FILE":   "value",
		"SHARED":      "flag",
		"EMPTY":       "",
		"WITH_EQUALS": "one=two",
	}, env)
}

func TestReadAdHocInputRequiresOneSource(t *testing.T) {
	_, err := readAdHocInput("echo hello", "script.sh")
	require.EqualError(t, err, "provide either --command or --script, not both")

	_, err = readAdHocInput("", "")
	require.EqualError(t, err, "either --command or --script is required")
}

func TestWaitForAdHocRunStreamsRawPaginatedLogs(t *testing.T) {
	api := &adHocAPI{
		runs: []*models.AppInstallActionWorkflowRun{{
			ID:     "run_123",
			Status: string(models.AppInstallActionWorkflowRunStatusFinished),
			RunnerJob: &models.AppRunnerJob{
				ID:          "job_123",
				LogStreamID: "logs_123",
			},
		}},
		logPages: []adHocLogPage{
			{
				records: []*models.AppOtelLogRecord{
					{ID: "plan_log", Body: "creating plan"},
					{ID: "lifecycle_log", RunnerJobID: "job_123", ScopeName: "oteljob", Body: "executing job"},
					{ID: "log_1", RunnerJobID: "job_123", Body: "first line", LogAttributes: map[string]string{"nuon.command_output": "true"}},
					{ID: "log_2", RunnerJobID: "job_123", Body: "second line\n", LogAttributes: map[string]string{"nuon.command_output": "true"}},
				},
				next:    "next-page",
				hasMore: true,
			},
			{
				records: []*models.AppOtelLogRecord{{ID: "log_3", RunnerJobID: "job_123", Body: "third line", LogAttributes: map[string]string{"nuon.command_output": "true"}}},
			},
			{},
		},
	}
	service := New(validator.New(), api, &config.Config{Viper: viper.New()})
	var output bytes.Buffer

	run, err := service.waitForAdHocRun(context.Background(), "inst_123", "run_123", &output)
	require.NoError(t, err)
	require.Equal(t, string(models.AppInstallActionWorkflowRunStatusFinished), run.Status)
	require.Equal(t, "first line\nsecond line\nthird line\n", output.String())
	require.Equal(t, []string{"", "next-page", "next-page", "next-page", "next-page", "next-page"}, api.logCursors)
}

func TestWaitForAdHocRunReturnsFailedTerminalStatus(t *testing.T) {
	api := &adHocAPI{runs: []*models.AppInstallActionWorkflowRun{{
		ID:                "run_123",
		Status:            string(models.AppInstallActionWorkflowRunStatusError),
		StatusDescription: "command failed",
	}}}
	service := New(validator.New(), api, &config.Config{Viper: viper.New()})

	run, err := service.waitForAdHocRun(context.Background(), "inst_123", "run_123", &bytes.Buffer{})
	require.NoError(t, err)
	require.Equal(t, string(models.AppInstallActionWorkflowRunStatusError), run.Status)
	require.Equal(t, "command failed", run.StatusDescription)
}

func TestWaitForAdHocRunHonorsCancellation(t *testing.T) {
	api := &adHocAPI{runs: []*models.AppInstallActionWorkflowRun{{
		ID:     "run_123",
		Status: string(models.AppInstallActionWorkflowRunStatusQueued),
	}}}
	service := New(validator.New(), api, &config.Config{Viper: viper.New()})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.waitForAdHocRun(ctx, "inst_123", "run_123", &bytes.Buffer{})
	require.ErrorIs(t, err, context.Canceled)
}

func TestUnknownAdHocStatusIsNotTerminal(t *testing.T) {
	require.False(t, isTerminalAdHocStatus(string(models.AppInstallActionWorkflowRunStatusUnknown)))
	require.True(t, isTerminalAdHocStatus(string(models.AppInstallActionWorkflowRunStatusTimedDashOut)))
}
