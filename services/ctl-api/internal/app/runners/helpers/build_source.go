package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func PersistBuildResult(ctx context.Context, db *gorm.DB, job *app.RunnerJob, result *app.RunnerJobExecutionResult, source *app.ComponentBuild) (*app.RunnerJobExecutionResult, bool, error) {
	var persisted *app.RunnerJobExecutionResult
	var created bool
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		persisted, created, err = CreateJobExecutionResultIfAbsent(ctx, tx, result)
		if err != nil || !created || !result.Success {
			return err
		}
		return applyBuildSource(tx, job, source, false)
	})
	return persisted, created, err
}

func ClaimBuildSource(ctx context.Context, db *gorm.DB, job *app.RunnerJob, executionID string, identity *sourceidentity.Identity) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var execution app.RunnerJobExecution
		if err := tx.Where(app.RunnerJobExecution{ID: executionID, RunnerJobID: job.ID}).First(&execution).Error; err != nil {
			return err
		}
		return applyBuildSource(tx, job, &app.ComponentBuild{SourceIdentity: identity}, true)
	})
}

func ResumeBuildSource(ctx context.Context, db *gorm.DB, jobID string, plan *plantypes.CompositePlan) (*plantypes.CompositePlan, error) {
	if plan.BuildPlan == nil || plan.BuildPlan.SourceIdentity == nil {
		return plan, nil
	}
	var job app.RunnerJob
	if err := db.WithContext(ctx).Select("owner_type", "owner_id").Where(app.RunnerJob{ID: jobID}).First(&job).Error; err != nil {
		return nil, err
	}
	var identity *sourceidentity.Identity
	switch job.OwnerType {
	case "component_builds":
		var build app.ComponentBuild
		if err := db.WithContext(ctx).Select("source_identity").Where(app.ComponentBuild{ID: job.OwnerID}).First(&build).Error; err != nil {
			return nil, err
		}
		identity = build.SourceIdentity
	case "app_sandbox_builds":
		var build app.AppSandboxBuild
		if err := db.WithContext(ctx).Select("source_identity").Where(app.AppSandboxBuild{ID: job.OwnerID}).First(&build).Error; err != nil {
			return nil, err
		}
		identity = build.SourceIdentity
	default:
		return nil, fmt.Errorf("source metadata requires a build owner")
	}
	if identity != nil && identity.Captured {
		if err := sourceidentity.ValidateTransition(plan.BuildPlan.SourceIdentity, identity, true); err != nil {
			return nil, err
		}
		plan.BuildPlan.SourceIdentity = identity
	}
	return plan, nil
}

func reportedBuildSource(expected, actual *sourceidentity.Identity) *sourceidentity.Identity {
	if actual != nil || expected == nil {
		return actual
	}
	unverified := *expected
	unverified.Files = append([]sourceidentity.File(nil), expected.Files...)
	unverified.Reasons = append([]string(nil), expected.Reasons...)
	unverified.MarkUnknown("runner_source_evidence_missing")
	unverified.Seal()
	return &unverified
}

func applyBuildSource(tx *gorm.DB, job *app.RunnerJob, source *app.ComponentBuild, claim bool) error {
	actual := source.SourceIdentity
	switch job.OwnerType {
	case "component_builds":
		var b app.ComponentBuild
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.ComponentBuild{ID: job.OwnerID}).First(&b).Error; err != nil {
			return err
		}
		if err := sourceidentity.ValidateTransition(b.SourceIdentity, actual, claim); err != nil {
			return err
		}
		actual = reportedBuildSource(b.SourceIdentity, actual)
		if actual != nil {
			b.SourceIdentity = actual
			b.Checksum = actual.ConfigChecksum
			if !claim {
				b.SourceChecksum = actual.EligibleChecksum()
			}
			if err := tx.Model(&b).Select("source_identity", "checksum", "source_checksum").Updates(&b).Error; err != nil {
				return err
			}
		}
		if source.SourceDigest != "" && !claim {
			if source.SourceIdentity != nil && actual.Digest != source.SourceDigest {
				return fmt.Errorf("image source digest mismatch")
			}
			return tx.Model(&b).Select("source_ref", "source_image", "resolved_tag", "source_digest", "source_media_type", "resolved_at", "no_op").Updates(source).Error
		}
	case "app_sandbox_builds":
		var b app.AppSandboxBuild
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.AppSandboxBuild{ID: job.OwnerID}).First(&b).Error; err != nil {
			return err
		}
		if err := sourceidentity.ValidateTransition(b.SourceIdentity, actual, claim); err != nil {
			return err
		}
		actual = reportedBuildSource(b.SourceIdentity, actual)
		if actual != nil {
			b.SourceIdentity = actual
			b.Checksum = actual.ConfigChecksum
			if !claim {
				b.SourceChecksum = actual.EligibleChecksum()
			}
			return tx.Model(&b).Select("source_identity", "checksum", "source_checksum").Updates(&b).Error
		}
	default:
		if actual != nil {
			return fmt.Errorf("source metadata requires a build owner")
		}
	}
	return nil
}
