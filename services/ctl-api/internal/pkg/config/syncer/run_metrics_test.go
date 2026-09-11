package syncer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	testseedconfig "github.com/nuonco/nuon/services/ctl-api/tests/testseed/config"
)

type runMetricsTemporalClient struct {
	temporal.Client
	err error
}

func (c *runMetricsTemporalClient) ExecuteWorkflowInNamespace(context.Context, string, tclient.StartWorkflowOptions, interface{}, ...interface{}) (tclient.WorkflowRun, error) {
	return &syncFieldsWorkflowRun{}, c.err
}

type RunMetricsSuite struct {
	tests.BaseDBTestSuite
	app  *fxtest.App
	deps syncDeps
	tc   *runMetricsTemporalClient
}

func TestRunMetrics(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
	}
	suite.Run(t, new(RunMetricsSuite))
}

func (s *RunMetricsSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	s.tc = &runMetricsTemporalClient{}
	options := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
		T:               s.T(),
		Mocks:           &tests.TestMocks{MockTC: s.tc},
		CustomValidator: true,
	}), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *RunMetricsSuite) TearDownSuite() { s.app.RequireStop() }

func (s *RunMetricsSuite) fixture(cfgJSON string) (context.Context, *app.App, *app.AppConfig) {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())
	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	appConfig := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	s.installBlobFixture(appConfig.ID, cfgJSON)
	return ctx, testApp, appConfig
}

func (s *RunMetricsSuite) installBlobFixture(id, value string) {
	name := "run_metrics_blob_" + id
	err := s.deps.DB.Callback().Query().After("gorm:query").Register(name, func(tx *gorm.DB) {
		loaded, ok := tx.Statement.Dest.(*app.AppConfig)
		if ok && loaded.ID == id {
			loaded.IntermediateConfig = &blobstore.Blob{}
			loaded.IntermediateConfig.Set(value)
		}
	})
	s.Require().NoError(err)
	s.T().Cleanup(func() { s.Require().NoError(s.deps.DB.Callback().Query().Remove(name)) })
}

func (s *RunMetricsSuite) run(ctx context.Context, testApp *app.App, appConfigID string) (*RunResult, error, map[attribute.Set]int64, map[string]interface{}) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	defer func() { s.Require().NoError(provider.Shutdown(context.Background())) }()
	core, logs := observer.New(zap.WarnLevel)

	result, err := Run(ctx, RunDeps{
		DB: s.deps.DB, AppsHelpers: s.deps.AppsHelpers, ComponentHelpers: s.deps.ComponentHelpers,
		ActionsHelpers: s.deps.ActionsHelpers, RunbooksHelpers: s.deps.RunbooksHelpers,
		InstallHelpers: s.deps.InstallHelpers, VCSHelpers: s.deps.VCSHelpers, TFClient: s.deps.TFClient,
		Metrics: NewMetrics(provider), Logger: zap.New(core),
	}, RunRequest{AppID: testApp.ID, AppConfigID: appConfigID})

	var data metricdata.ResourceMetrics
	s.Require().NoError(reader.Collect(context.Background(), &data))
	var attempts map[attribute.Set]int64
	var histogramCount uint64
	for _, scope := range data.ScopeMetrics {
		for _, measurement := range scope.Metrics {
			switch measurement.Name {
			case "nuon.config.sync.attempts":
				attempts = int64Points(measurement.Data.(metricdata.Sum[int64]).DataPoints)
			case "nuon.config.sync.duration":
				for _, point := range measurement.Data.(metricdata.Histogram[float64]).DataPoints {
					histogramCount += point.Count
				}
			}
		}
	}
	s.EqualValues(1, histogramCount)
	fields := map[string]interface{}{}
	if logs.Len() != 0 {
		fields = logs.All()[0].ContextMap()
	}
	return result, err, attempts, fields
}

func (s *RunMetricsSuite) assertStatus(id string, status app.AppConfigStatus) {
	var persisted app.AppConfig
	s.Require().NoError(s.deps.DB.First(&persisted, &app.AppConfig{ID: id}).Error)
	s.Equal(status, persisted.Status)
	if status == app.AppConfigStatusError {
		s.Contains(persisted.StatusDescription, "sync failed:")
	}
}

func (s *RunMetricsSuite) TestSuccessAndRepeatedAttempts() {
	encoded, err := json.Marshal(testseedconfig.BuildMinimalAppConfig())
	s.Require().NoError(err)
	ctx, testApp, appConfig := s.fixture(string(encoded))
	for range 2 {
		result, runErr, points, fields := s.run(ctx, testApp, appConfig.ID)
		s.Require().NoError(runErr)
		s.Require().NotNil(result)
		s.Equal(map[attribute.Set]int64{attribute.NewSet(attribute.String("outcome", "success"), attribute.String("stage", "none")): 1}, points)
		s.Empty(fields)
		s.assertStatus(appConfig.ID, app.AppConfigStatusActive)
	}
}

func (s *RunMetricsSuite) TestMissingConfig() {
	ctx, testApp, _ := s.fixture("{}")
	_, err, points, fields := s.run(ctx, testApp, "cfg00000000000000000000000")
	s.Error(err)
	s.Equal(int64(1), points[attribute.NewSet(attribute.String("outcome", "error"), attribute.String("stage", "load"))])
	s.Equal(false, fields["config_committed"])
}

func (s *RunMetricsSuite) TestInvalidJSON() {
	ctx, testApp, appConfig := s.fixture("{")
	_, err, points, fields := s.run(ctx, testApp, appConfig.ID)
	s.ErrorContains(err, "unable to unmarshal intermediate config")
	s.Equal(int64(1), points[attribute.NewSet(attribute.String("outcome", "error"), attribute.String("stage", "decode"))])
	s.Equal(false, fields["config_committed"])
	s.assertStatus(appConfig.ID, app.AppConfigStatusError)
}

type runMetricsCommitFailurePool struct{ gorm.ConnPool }

func (p runMetricsCommitFailurePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	tx, err := p.ConnPool.(gorm.TxBeginner).BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &runMetricsCommitFailureTx{tx}, nil
}

type runMetricsCommitFailureTx struct{ *sql.Tx }

func (tx runMetricsCommitFailureTx) Commit() error {
	_ = tx.Rollback()
	return errors.New("injected commit failure")
}

func (s *RunMetricsSuite) TestCommitFailure() {
	encoded, err := json.Marshal(testseedconfig.BuildMinimalAppConfig())
	s.Require().NoError(err)
	ctx, testApp, appConfig := s.fixture(string(encoded))
	db := s.deps.DB
	s.deps.DB = db.Session(&gorm.Session{Context: ctx, NewDB: true, SkipDefaultTransaction: true})
	s.deps.DB.Statement.ConnPool = runMetricsCommitFailurePool{db.Statement.ConnPool}
	defer func() { s.deps.DB = db }()

	_, runErr, points, fields := s.run(ctx, testApp, appConfig.ID)
	s.ErrorContains(runErr, "injected commit failure")
	s.Equal(map[attribute.Set]int64{attribute.NewSet(attribute.String("outcome", "error"), attribute.String("stage", "sync_transaction")): 1}, points)
	s.Equal(false, fields["config_committed"])
	s.assertStatus(appConfig.ID, app.AppConfigStatusError)
	var persisted app.AppConfig
	s.Require().NoError(db.Where(app.AppConfig{ID: appConfig.ID}).First(&persisted).Error)
	s.Empty(persisted.State)
}

func (s *RunMetricsSuite) TestCancellation() {
	ctx, testApp, appConfig := s.fixture("{}")
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	_, runErr, points, fields := s.run(ctx, testApp, appConfig.ID)
	s.ErrorIs(runErr, context.Canceled)
	s.Equal(map[attribute.Set]int64{attribute.NewSet(attribute.String("outcome", "cancelled"), attribute.String("stage", "load")): 1}, points)
	s.Equal(false, fields["config_committed"])
}

func (s *RunMetricsSuite) TestPostCommitQueueFailure() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Components = config.ComponentList{terraformComponent("queue-failure")}
	encoded, err := json.Marshal(cfg)
	s.Require().NoError(err)
	ctx, testApp, appConfig := s.fixture(string(encoded))
	s.tc.err = errors.New("injected queue startup failure")
	defer func() { s.tc.err = nil }()

	_, runErr, points, fields := s.run(ctx, testApp, appConfig.ID)
	s.ErrorContains(runErr, "injected queue startup failure")
	s.Equal(map[attribute.Set]int64{attribute.NewSet(attribute.String("outcome", "error"), attribute.String("stage", "deferred_queues")): 1}, points)
	s.Equal(true, fields["config_committed"])
	s.assertStatus(appConfig.ID, app.AppConfigStatusError)
	var component app.Component
	s.Require().NoError(s.deps.DB.Where(app.Component{AppID: testApp.ID, Name: "queue-failure"}).First(&component).Error)
	var persisted app.AppConfig
	s.Require().NoError(s.deps.DB.Where(app.AppConfig{ID: appConfig.ID}).First(&persisted).Error)
	s.Contains([]string(persisted.ComponentIDs), component.ID)
}

func (s *RunMetricsSuite) TestVerifiedValidationRejection() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cfg.Components = config.ComponentList{testseedconfig.BuildDockerBuildComponent("unsupported")}
	encoded, err := json.Marshal(cfg)
	s.Require().NoError(err)
	ctx, testApp, appConfig := s.fixture(string(encoded))
	_, runErr, points, fields := s.run(ctx, testApp, appConfig.ID)
	s.Error(runErr)
	s.Equal(int64(1), points[attribute.NewSet(attribute.String("outcome", "rejected"), attribute.String("stage", "sync_transaction"))], "error: %v; points: %#v", runErr, points)
	s.Equal(false, fields["config_committed"])
	s.assertStatus(appConfig.ID, app.AppConfigStatusError)
}

func (s *RunMetricsSuite) TestTransactionInfrastructureFailure() {
	encoded, err := json.Marshal(testseedconfig.BuildMinimalAppConfig())
	s.Require().NoError(err)
	ctx, testApp, appConfig := s.fixture(string(encoded))
	callback := "run_metrics_transaction_failure_" + appConfig.ID
	s.Require().NoError(s.deps.DB.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "app_sandbox_configs" {
			tx.AddError(errors.New("injected transaction failure"))
		}
	}))
	defer func() { s.Require().NoError(s.deps.DB.Callback().Create().Remove(callback)) }()

	_, runErr, points, fields := s.run(ctx, testApp, appConfig.ID)
	s.ErrorContains(runErr, "injected transaction failure")
	s.Equal(int64(1), points[attribute.NewSet(attribute.String("outcome", "error"), attribute.String("stage", "sync_transaction"))])
	s.Equal(false, fields["config_committed"])
	s.assertStatus(appConfig.ID, app.AppConfigStatusError)
}
