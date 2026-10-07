package activities

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/oci/signature"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type updateBuildSignatureVerificationTestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestUpdateBuildSignatureVerificationSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(updateBuildSignatureVerificationTestSuite))
}

func (s *updateBuildSignatureVerificationTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	cfg, err := tests.LoadDBConfig()
	require.NoError(s.T(), err)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	s.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(s.T(), err)
	s.seeder = testseed.New(testseed.Params{DB: s.db})
	s.SetDB(s.db)
}

func (s *updateBuildSignatureVerificationTestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

func (s *updateBuildSignatureVerificationTestSuite) seedBuild() (context.Context, *app.ComponentBuild, *app.RunnerJob) {
	t := s.T()
	ctx, _ := s.seeder.EnsureAccount(context.Background(), t)
	ctx, _ = s.seeder.EnsureOrg(ctx, t)

	a := s.seeder.CreateApp(ctx, t)
	cfg := s.seeder.CreateAppConfig(ctx, t, a.ID)
	comp := s.seeder.CreateComponent(ctx, t, a.ID, app.ComponentTypeExternalImage)
	ccc := s.seeder.CreateExternalImageComponentConfigConnection(ctx, t, comp.ID, cfg.ID)
	bld := s.seeder.CreateComponentBuild(ctx, t, ccc.ID)
	job := s.seeder.CreateRunnerJob(ctx, t, bld.ID, "component_builds")
	return ctx, bld, job
}

func (s *updateBuildSignatureVerificationTestSuite) addExecution(ctx context.Context, jobID string, createdAt time.Time, step *string) {
	t := s.T()
	exec := &app.RunnerJobExecution{
		RunnerJobID: jobID,
		Status:      app.RunnerJobExecutionStatusFailed,
		CreatedAt:   createdAt,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(exec).Error)
	if step == nil {
		return
	}
	require.NoError(t, s.db.WithContext(ctx).Create(&app.RunnerJobExecutionResult{
		RunnerJobExecutionID: exec.ID,
		ErrorMetadata:        pgtype.Hstore{"step": step},
	}).Error)
}

func (s *updateBuildSignatureVerificationTestSuite) run(ctx context.Context, bld *app.ComponentBuild, job *app.RunnerJob, required, succeeded bool) app.ComponentBuildSignatureVerification {
	t := s.T()
	acts := &Activities{db: s.db}
	require.NoError(t, acts.UpdateBuildSignatureVerification(ctx, &UpdateBuildSignatureVerificationRequest{
		BuildID:      bld.ID,
		JobID:        job.ID,
		Required:     required,
		JobSucceeded: succeeded,
	}))

	var got app.ComponentBuild
	require.NoError(t, s.db.WithContext(ctx).First(&got, "id = ?", bld.ID).Error)
	return got.SignatureVerification
}

func strPtr(s string) *string { return &s }

func (s *updateBuildSignatureVerificationTestSuite) TestNotRequired() {
	for _, succeeded := range []bool{true, false} {
		ctx, bld, job := s.seedBuild()
		s.addExecution(ctx, job.ID, time.Now(), strPtr(signature.VerifyStep))
		require.Equal(s.T(), app.ComponentBuildSignatureVerificationNotRequired, s.run(ctx, bld, job, false, succeeded))
	}
}

func (s *updateBuildSignatureVerificationTestSuite) TestVerified() {
	ctx, bld, job := s.seedBuild()
	require.Equal(s.T(), app.ComponentBuildSignatureVerificationVerified, s.run(ctx, bld, job, true, true))
}

func (s *updateBuildSignatureVerificationTestSuite) TestRejected() {
	ctx, bld, job := s.seedBuild()
	s.addExecution(ctx, job.ID, time.Now(), strPtr(signature.VerifyStep))
	require.Equal(s.T(), app.ComponentBuildSignatureVerificationRejected, s.run(ctx, bld, job, true, false))
}

func (s *updateBuildSignatureVerificationTestSuite) TestOtherFailureLeavesEmpty() {
	ctx, bld, job := s.seedBuild()
	s.addExecution(ctx, job.ID, time.Now(), strPtr("copy image"))
	require.Empty(s.T(), s.run(ctx, bld, job, true, false))
}

func (s *updateBuildSignatureVerificationTestSuite) TestNoExecutionLeavesEmpty() {
	ctx, bld, job := s.seedBuild()
	require.Empty(s.T(), s.run(ctx, bld, job, true, false))
}

func (s *updateBuildSignatureVerificationTestSuite) TestExecutionWithoutResultLeavesEmpty() {
	ctx, bld, job := s.seedBuild()
	s.addExecution(ctx, job.ID, time.Now(), nil)
	require.Empty(s.T(), s.run(ctx, bld, job, true, false))
}

func (s *updateBuildSignatureVerificationTestSuite) TestUsesLatestExecution() {
	now := time.Now()

	ctx, bld, job := s.seedBuild()
	s.addExecution(ctx, job.ID, now.Add(-time.Minute), strPtr("copy image"))
	s.addExecution(ctx, job.ID, now, strPtr(signature.VerifyStep))
	require.Equal(s.T(), app.ComponentBuildSignatureVerificationRejected, s.run(ctx, bld, job, true, false))

	ctx, bld, job = s.seedBuild()
	s.addExecution(ctx, job.ID, now.Add(-time.Minute), strPtr(signature.VerifyStep))
	s.addExecution(ctx, job.ID, now, strPtr("copy image"))
	require.Empty(s.T(), s.run(ctx, bld, job, true, false))
}
