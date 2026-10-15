package activities_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type installStateIntegrationDeps struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	Cfg            *internal.Config
	Helpers        *helpers.Helpers
	Seeder         *testseed.Seeder
	Blobs          blobstore.Service
	AppsHelpers    *appshelpers.Helpers
	ActionHelpers  *actionshelpers.Helpers
	RunnersHelpers *runnershelpers.Helpers
	Audit          *audit.Emitter
}

type InstallStateIntegrationSuite struct {
	tests.BaseDBTestSuite

	fxApp      *fxtest.App
	deps       installStateIntegrationDeps
	activities *installactivities.Activities
	s3Client   *s3.Client
	bucket     string
	ctx        context.Context
	install    *app.Install
	inputCfg   *app.AppInputConfig
	reader     *sdkmetric.ManualReader
}

type failHeadUpdate struct{}

func TestInstallStateIntegrationSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	if os.Getenv("AWS_ENDPOINT_URL_S3") == "" {
		t.Skip("set AWS_ENDPOINT_URL_S3 to the local nctl S3 service")
	}
	suite.Run(t, new(InstallStateIntegrationSuite))
}

func (s *InstallStateIntegrationSuite) SetupSuite() {
	endpoint, err := url.Parse(os.Getenv("AWS_ENDPOINT_URL_S3"))
	require.NoError(s.T(), err)
	require.Contains(s.T(), []string{"127.0.0.1", "localhost", "::1"}, endpoint.Hostname())
	s.T().Setenv("AWS_ACCESS_KEY_ID", "test")
	s.T().Setenv("AWS_SECRET_ACCESS_KEY", "test")
	s.T().Setenv("AWS_SESSION_TOKEN", "")
	s.T().Setenv("AWS_REGION", "us-east-1")
	s.T().Setenv("AWS_EC2_METADATA_DISABLED", "true")

	s.bucket = "nuon-state-repro-" + domains.NewBlobID()
	mockTC := temporal.NewMockClient(gomock.NewController(s.T()))
	s.reader = sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(s.reader))
	s.T().Cleanup(func() { require.NoError(s.T(), provider.Shutdown(context.Background())) })
	opts := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
		T: s.T(), CustomValidator: true,
		Mocks: &tests.TestMocks{MockTC: mockTC},
	}),
		fx.Provide(func() metric.MeterProvider { return provider }),
		fx.Decorate(func(cfg *internal.Config) *internal.Config {
			require.Contains(s.T(), []string{"127.0.0.1", "localhost", "::1"}, cfg.DBHost)
			require.NotEqual(s.T(), "ctl_api", cfg.DBName)
			cfg.BlobStorageProvider = "s3"
			cfg.BlobStorageBucket = s.bucket
			cfg.BlobStorageRegion = "us-east-1"
			cfg.BlobReadEnabled = true
			return cfg
		}),
		fx.Populate(&s.deps),
	)
	s.fxApp = fxtest.New(s.T(), opts...)
	require.NoError(s.T(), s.deps.DB.Callback().Update().Before("gorm:update").Register("test:fail_install_state_head", func(tx *gorm.DB) {
		if tx.Statement.Table == "install_state_heads" && tx.Statement.Context.Value(failHeadUpdate{}) != nil {
			_ = tx.AddError(fmt.Errorf("injected head update failure"))
		}
	}))
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
	s.activities = installactivities.New(installactivities.Params{
		DB: s.deps.DB, Cfg: s.deps.Cfg, Helpers: s.deps.Helpers, AppsHelpers: s.deps.AppsHelpers,
		ActionHelpers: s.deps.ActionHelpers, RunnersHelpers: s.deps.RunnersHelpers, Audit: s.deps.Audit,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	require.NoError(s.T(), err)
	s.s3Client = s3.NewFromConfig(cfg, func(opts *s3.Options) {
		opts.BaseEndpoint = aws.String(endpoint.String())
		opts.UsePathStyle = true
	})
	_, err = s.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)})
	require.NoError(s.T(), err)
}

func (s *InstallStateIntegrationSuite) SetupTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	s.T().Cleanup(cancel)
	ctx = blobstore.WithBlobService(ctx, s.deps.Blobs)
	ctx = blobstore.WithBlobWriteEnabled(ctx, true)
	ctx, _ = s.deps.Seeder.EnsureAccount(ctx, s.T())
	s.ctx, _ = s.deps.Seeder.EnsureOrg(ctx, s.T())
	s.deps.Cfg.BlobReadEnabled = true
	application := s.deps.Seeder.CreateApp(s.ctx, s.T())
	appCfg := s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), application.ID)
	s.install = s.deps.Seeder.CreateInstall(s.ctx, s.T(), application)
	s.inputCfg = &app.AppInputConfig{}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.AppInputConfig{AppConfigID: appCfg.ID}).First(s.inputCfg).Error)
	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), s.install.ID, s.inputCfg.ID,
		map[string]*string{"region": generics.ToPtr("us-east-1")})

	require.Equal(s.T(), "us-east-1", s.fresh().Inputs.Inputs["region"])
	stored, fromBlob := s.currentRow().GetState(s.ctx, true)
	require.True(s.T(), fromBlob, "the fixture must round-trip through real S3, not the inline fallback")
	require.Equal(s.T(), "us-east-1", stored.Inputs.Inputs["region"])
}

func (s *InstallStateIntegrationSuite) TearDownSuite() {
	s.fxApp.RequireStop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pages := s3.NewListObjectsV2Paginator(s.s3Client, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket)})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		require.NoError(s.T(), err)
		for _, object := range page.Contents {
			_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: object.Key})
			require.NoError(s.T(), err)
		}
	}
	_, err := s.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(s.bucket)})
	require.NoError(s.T(), err)
}

func (s *InstallStateIntegrationSuite) fresh() *state.State {
	s.T().Helper()
	snapshot, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Nil(s.T(), snapshot.StaleAt)
	return snapshot
}

func (s *InstallStateIntegrationSuite) head() app.InstallStateHead {
	s.T().Helper()
	var head app.InstallStateHead
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallStateHead{InstallID: s.install.ID}).First(&head).Error)
	return head
}

func (s *InstallStateIntegrationSuite) currentRow() *app.InstallState {
	s.T().Helper()
	row := &app.InstallState{}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallState{ID: s.head().CurrentStateID}).First(row).Error)
	return row
}

func (s *InstallStateIntegrationSuite) snapshotCount() int64 {
	s.T().Helper()
	var count int64
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Scopes(scopes.WithDisableViews).Model(&app.InstallState{}).
		Where(app.InstallState{InstallID: s.install.ID}).Count(&count).Error)
	return count
}

func (s *InstallStateIntegrationSuite) baseline() *helpers.InstallStateBaseline {
	s.T().Helper()
	baseline, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, s.install.ID)
	require.NoError(s.T(), err)
	return baseline
}

func (s *InstallStateIntegrationSuite) publish(ctx context.Context, baseline *helpers.InstallStateBaseline, refreshed ...pkgstate.PartialName) (*app.InstallState, error) {
	return s.deps.Helpers.PublishInstallState(ctx, &app.InstallState{
		InstallID: s.install.ID, State: baseline.State, GeneratedBy: app.InstallStateGenerateSourceStateManager,
		TriggeredByID: s.install.ID, TriggeredByType: "installs",
	}, baseline.Head, refreshed)
}

func (s *InstallStateIntegrationSuite) updateInput() {
	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), s.install.ID, s.inputCfg.ID,
		map[string]*string{"region": generics.ToPtr("eu-west-1")})
}

func (s *InstallStateIntegrationSuite) renameApp(name string) error {
	return s.deps.DB.WithContext(s.ctx).Model(&app.App{ID: s.install.AppID}).Update("name", name).Error
}

type publicationBlobService struct {
	blobstore.Service
	afterUpload func() error
}

func (b publicationBlobService) UploadStream(ctx context.Context, key string, data io.Reader) (string, error) {
	checksum, err := b.Service.UploadStream(ctx, key, data)
	if err == nil {
		err = b.afterUpload()
	}
	return checksum, err
}

func (s *InstallStateIntegrationSuite) withUploadHook(hook func() error) context.Context {
	return blobstore.WithBlobService(s.ctx, publicationBlobService{Service: s.deps.Blobs, afterUpload: hook})
}

func (s *InstallStateIntegrationSuite) TestPartialRefreshPublishesImmutableBlobSnapshots() {
	original := s.currentRow()
	s.updateInput()
	snapshot := s.fresh()
	require.Equal(s.T(), "eu-west-1", snapshot.Inputs.Inputs["region"])
	require.Equal(s.T(), "eu-west-1", snapshot.Install.Inputs["region"])
	refreshed := s.head()
	require.NotEqual(s.T(), original.ID, refreshed.CurrentStateID)
	require.Empty(s.T(), refreshed.StalePartials)

	snapshot = s.fresh()
	require.Equal(s.T(), "eu-west-1", snapshot.Inputs.Inputs["region"], "a repeated read must not revert to the old blob")
	require.Equal(s.T(), refreshed, s.head(), "a clean head must be read without republishing")

	stored, fromBlob := original.GetState(s.ctx, true)
	require.True(s.T(), fromBlob)
	require.Equal(s.T(), "us-east-1", stored.Inputs.Inputs["region"], "published snapshots must remain immutable")

	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallState{}).
		Where(app.InstallState{ID: refreshed.CurrentStateID}).Select("state").
		Updates(map[string]any{"state": nil}).Error)
	require.Equal(s.T(), "eu-west-1", s.fresh().Inputs.Inputs["region"], "the refreshed snapshot must be readable from its blob alone")

	readCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	readCtx.Request = httptest.NewRequest("GET", "/v1/installs/"+s.install.ID+"/state", nil).WithContext(s.ctx)
	orgID, err := cctx.OrgIDFromContext(s.ctx)
	require.NoError(s.T(), err)
	cctx.SetOrgIDGinContext(readCtx, orgID)
	accountID, err := cctx.AccountIDFromContext(s.ctx)
	require.NoError(s.T(), err)
	cctx.SetAccountIDGinContext(readCtx, accountID)
	for read := 0; read < 2; read++ {
		snapshot, err := s.deps.Helpers.GetInstallState(readCtx, s.install.ID, true, true)
		require.NoError(s.T(), err)
		require.Equal(s.T(), "eu-west-1", snapshot.Inputs.Inputs["region"])
		require.Equal(s.T(), snapshot.Inputs.Inputs, snapshot.Install.Inputs)
	}
	require.Equal(s.T(), refreshed, s.head())

	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), s.install.ID, s.inputCfg.ID,
		map[string]*string{"region": generics.ToPtr("ap-south-1")})
	saved, err := s.activities.SaveState(s.ctx, &installactivities.SaveStateRequest{
		State: stored, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
		GeneratedBy: app.InstallStateGenerateSourceStateManager, RefreshedPartials: []pkgstate.PartialName{pkgstate.PartialInputs},
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), saved.ID, s.head().CurrentStateID)
	for read := 0; read < 2; read++ {
		snapshot, err := s.deps.Helpers.GetInstallState(readCtx, s.install.ID, true, true)
		require.NoError(s.T(), err)
		require.Nil(s.T(), snapshot.StaleAt)
		require.Equal(s.T(), "ap-south-1", snapshot.Inputs.Inputs["region"], "legacy SaveState must ignore its cached payload")
		require.Equal(s.T(), snapshot.Inputs.Inputs, snapshot.Install.Inputs)
	}
}

func (s *InstallStateIntegrationSuite) TestRedactedRefreshPreservesCanonicalInputsAndDefaults() {
	older := s.baseline()
	var region app.AppInput
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.AppInput{AppInputConfigID: s.inputCfg.ID, Name: "region"}).First(&region).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&region).Update("sensitive", true).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppInput{
		AppInputConfigID: s.inputCfg.ID, AppInputGroupID: region.AppInputGroupID,
		Name: "replicas", DisplayName: "Replicas", Description: "Replica count",
		Type: app.AppInputTypeString, Default: "3",
	}).Error)
	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), s.install.ID, s.inputCfg.ID,
		map[string]*string{
			"region": generics.ToPtr("example-secret"), "replicas": generics.ToPtr(""),
			"undeclared": generics.ToPtr("ignore-me"),
		})
	for read := 0; read < 2; read++ {
		redacted, err := s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, true, true)
		require.NoError(s.T(), err)
		require.Nil(s.T(), redacted.StaleAt)
		require.Equal(s.T(), map[string]string{"region": "********", "replicas": "3"}, redacted.Inputs.Inputs)
		require.Equal(s.T(), redacted.Inputs.Inputs, redacted.Install.Inputs)
	}
	canonical := map[string]string{"region": "example-secret", "replicas": "3"}
	require.Equal(s.T(), canonical, s.fresh().Inputs.Inputs)
	stored, _ := s.currentRow().GetState(s.ctx, true)
	require.Equal(s.T(), canonical, stored.Inputs.Inputs, "redaction must never be persisted")

	_, err := s.publish(s.ctx, older, pkgstate.PartialInputs)
	require.True(s.T(), helpers.IsInstallStateConflict(err), "%v", err)
	require.Equal(s.T(), canonical, s.fresh().Inputs.Inputs)
}

func (s *InstallStateIntegrationSuite) TestLegacySnapshotsBackfilledBeforeReads() {
	existing := s.head()
	for name, tc := range map[string]struct {
		blobOnly   bool
		noSnapshot bool
		deleted    bool
	}{
		"divergent_inline_and_blob":    {},
		"archived_blob_reads_disabled": {blobOnly: true},
		"no_snapshot":                  {noSnapshot: true},
		"deleted_install":              {deleted: true},
	} {
		s.Run(name, func() {
			install := s.deps.Seeder.CreateInstall(s.ctx, s.T(), &app.App{ID: s.install.AppID})
			s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), install.ID, s.inputCfg.ID,
				map[string]*string{"region": generics.ToPtr("eu-west-1")})
			row := &app.InstallState{InstallID: install.ID, State: &state.State{
				ID: install.ID, Name: "legacy-snapshot", Inputs: &state.InputsState{Inputs: map[string]string{"region": "us-east-1"}},
			}}
			if !tc.noSnapshot {
				encoded, err := json.Marshal(row.State)
				require.NoError(s.T(), err)
				row.StateBlob = &blobstore.Blob{}
				row.StateBlob.Set(string(encoded))
				s.create(row)
				if tc.blobOnly {
					row.State, row.Archived = nil, true
				} else {
					row.State.Inputs.Inputs["region"] = "eu-west-1"
				}
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(row).Select("state", "archived").Updates(row).Error)
				stored, fromBlob := row.GetState(s.ctx, true)
				require.True(s.T(), fromBlob)
				require.Equal(s.T(), "us-east-1", stored.Inputs.Inputs["region"])
			}
			if tc.deleted {
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(install).Error)
			}
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Delete(&app.InstallStateHead{InstallID: install.ID}).Error)
			migrate := new(psqlmigrations.Migrations).Migration146InstallStateHeads
			require.NoError(s.T(), migrate(s.ctx, s.deps.DB))
			baseline, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, install.ID)
			if tc.deleted {
				require.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)
				require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Unscoped().Model(&app.Install{ID: install.ID}).Update("deleted_at", 0).Error)
				baseline, err = s.deps.Helpers.GetInstallStateBaseline(s.ctx, install.ID)
			}
			require.NoError(s.T(), err)
			require.Equal(s.T(), int64(1), baseline.Head.Revision)
			require.Equal(s.T(), install.OrgID, baseline.Head.OrgID)
			require.Empty(s.T(), baseline.Head.CurrentStateID)
			require.False(s.T(), baseline.Head.StaleAt.Empty())
			require.Equal(s.T(), pkgstate.AllPartials, baseline.Head.StalePartials)
			s.deps.Cfg.BlobReadEnabled = !tc.blobOnly
			for range 2 {
				snapshot, err := s.deps.Helpers.GetInstallState(s.ctx, install.ID, false, true)
				require.NoError(s.T(), err)
				require.Nil(s.T(), snapshot.StaleAt)
				require.Equal(s.T(), "eu-west-1", snapshot.Inputs.Inputs["region"])
			}
			s.deps.Cfg.BlobReadEnabled = true
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: install.ID}).Update("name", "after-"+install.ID).Error)
			pending, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, install.ID)
			require.NoError(s.T(), err)
			require.NoError(s.T(), migrate(s.ctx, s.deps.DB))
			retried, err := s.deps.Helpers.GetInstallStateBaseline(s.ctx, install.ID)
			require.NoError(s.T(), err)
			require.Equal(s.T(), pending.Head, retried.Head, "rerunning the migration must preserve advanced heads")
			require.Equal(s.T(), existing, s.head())
		})
	}
}

func (s *InstallStateIntegrationSuite) TestFailedPublicationLeavesHeadStale() {
	s.updateInput()
	pending := s.head()
	require.False(s.T(), pending.StaleAt.Empty())
	for name, ctx := range map[string]context.Context{
		"write_disabled": blobstore.WithBlobWriteEnabled(s.ctx, false),
		"upload_failed":  s.withUploadHook(func() error { return fmt.Errorf("injected blob upload failure") }),
		"head_update_failed_after_snapshot_insert": context.WithValue(s.ctx, failHeadUpdate{}, true),
	} {
		s.Run(name, func() {
			before := s.snapshotCount()
			_, err := s.activities.GetInstallState(ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
			require.Error(s.T(), err)
			require.Equal(s.T(), pending, s.head())
			require.Equal(s.T(), before, s.snapshotCount(), "a failed publication must not leave a snapshot row")
			_, err = s.activities.SaveState(ctx, &installactivities.SaveStateRequest{
				State: s.baseline().State, InstallID: s.install.ID, TriggeredByID: s.install.ID, TriggeredByType: "installs",
				GeneratedBy: app.InstallStateGenerateSourceStateManager, RefreshedPartials: []pkgstate.PartialName{pkgstate.PartialInputs},
			})
			require.Error(s.T(), err)
			require.Equal(s.T(), pending, s.head())
			require.Equal(s.T(), before, s.snapshotCount())
			require.Equal(s.T(), "us-east-1", s.baseline().State.Inputs.Inputs["region"])
		})
	}
	for read := 0; read < 2; read++ {
		require.Equal(s.T(), "eu-west-1", s.fresh().Inputs.Inputs["region"])
	}
}
