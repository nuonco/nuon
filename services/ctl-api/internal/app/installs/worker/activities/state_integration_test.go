package activities_test

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/metrics"
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
	MW             metrics.Writer
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
}

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
	opts := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
		T: s.T(), CustomValidator: true,
		Mocks: &tests.TestMocks{MockTC: mockTC},
	}),
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
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

	snapshot := &state.State{
		ID: s.install.ID, Name: s.install.Name,
		Inputs: &state.InputsState{Inputs: map[string]string{"region": "us-east-1"}},
	}
	helpers.MapLegacyFields(snapshot)
	row, err := s.saveState(snapshot)
	require.NoError(s.T(), err)
	var reloaded app.InstallState
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallState{ID: row.ID}).First(&reloaded).Error)
	stored, fromBlob := reloaded.GetState(s.ctx, true)
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

func (s *InstallStateIntegrationSuite) saveState(snapshot *state.State) (*app.InstallState, error) {
	return s.activities.SaveState(s.ctx, &installactivities.SaveStateRequest{
		State: snapshot, InstallID: s.install.ID,
		TriggeredByID: s.install.ID, TriggeredByType: "installs",
		GeneratedBy: app.InstallStateGenerateSourceStateManager,
	})
}

func (s *InstallStateIntegrationSuite) updateInput() {
	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), s.install.ID, s.inputCfg.ID,
		map[string]*string{"region": generics.ToPtr("eu-west-1")})
	require.NoError(s.T(), s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, pkgstate.PartialInputs))
}

func (s *InstallStateIntegrationSuite) TestPartialRefreshRemainsFreshOnRepeatedRead() {
	var original app.InstallState
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallState{InstallID: s.install.ID}).Order("created_at DESC").First(&original).Error)
	s.updateInput()
	for read := 1; read <= 2; read++ {
		snapshot, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
		require.NoError(s.T(), err)
		require.Equal(s.T(), "eu-west-1", snapshot.Inputs.Inputs["region"], "read %d must not revert to the old S3 blob", read)
		require.Equal(s.T(), "eu-west-1", snapshot.Install.Inputs["region"], "legacy input reference on read %d", read)
	}
	stored, fromBlob := original.GetState(s.ctx, true)
	require.True(s.T(), fromBlob)
	require.Equal(s.T(), "us-east-1", stored.Inputs.Inputs["region"], "an existing blob reference must remain immutable")
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallState{}).
		Where(app.InstallState{ID: original.ID}).Select("state", "archived").
		Updates(app.InstallState{State: nil, Archived: true}).Error)
	snapshot, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), "eu-west-1", snapshot.Inputs.Inputs["region"], "refreshed state must remain readable without its inline copy")
}

func (s *InstallStateIntegrationSuite) TestLateGenerationDoesNotRestoreOldInputs() {
	s.deps.Cfg.BlobReadEnabled = false
	older, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), "us-east-1", older.Inputs.Inputs["region"])
	s.updateInput()
	current, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), "eu-west-1", current.Inputs.Inputs["region"])
	_, err = s.saveState(current)
	require.NoError(s.T(), err)
	if _, err = s.saveState(older); err != nil {
		s.T().Logf("late snapshot save rejected: %v", err)
	}
	latest, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), "eu-west-1", latest.Inputs.Inputs["region"], "a late generation must not restore an older input value")
	s.deps.Cfg.BlobReadEnabled = true
	latest, err = s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), "eu-west-1", latest.Inputs.Inputs["region"])
	require.Equal(s.T(), "eu-west-1", latest.Install.Inputs["region"])
}

func (s *InstallStateIntegrationSuite) TestRedactedRefreshPreservesCanonicalInputsAndDefaults() {
	older, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
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
	require.NoError(s.T(), s.deps.Helpers.MarkInstallStatePartialsStale(s.ctx, s.deps.DB, s.install.ID, pkgstate.PartialInputs))
	for read := 0; read < 2; read++ {
		redacted, err := s.deps.Helpers.GetInstallState(s.ctx, s.install.ID, true, true)
		require.NoError(s.T(), err)
		require.Equal(s.T(), map[string]string{"region": "********", "replicas": "3"}, redacted.Inputs.Inputs)
		require.Equal(s.T(), redacted.Inputs.Inputs, redacted.Install.Inputs)
	}
	canonical, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), map[string]string{"region": "example-secret", "replicas": "3"}, canonical.Inputs.Inputs)
	_, err = s.saveState(older)
	require.NoError(s.T(), err)
	canonical, err = s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.NoError(s.T(), err)
	require.Equal(s.T(), map[string]string{"region": "example-secret", "replicas": "3"}, canonical.Inputs.Inputs)
}

func (s *InstallStateIntegrationSuite) TestFailedBlobPublicationLeavesInputsStale() {
	s.updateInput()
	_, err := s.activities.GetInstallState(blobstore.WithBlobWriteEnabled(s.ctx, false), &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
	require.Error(s.T(), err)
	for read := 0; read < 2; read++ {
		snapshot, err := s.activities.GetInstallState(s.ctx, &installactivities.GetInstallStateRequest{InstallID: s.install.ID})
		require.NoError(s.T(), err)
		require.Equal(s.T(), "eu-west-1", snapshot.Inputs.Inputs["region"])
	}
}
