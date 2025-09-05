package activities

import (
	"context"
)

type SeedImportDBResponse struct {
	Size string `json:"size" validate:"required"`
}

type SeedImportDB struct {
	BackupFP       string `json:"backup_fp" validate:"required"`
	BackupS3Bucket string `json:"backup_s3_bucket" validate:"required"`
	BackupIAMRole  string `json:"backup_iam_role" validate:"required"`
}

// @temporal-gen activity
func (a *Activities) SeedImportDB(ctx context.Context, req SeedImportDB) (*SeedImportDBResponse, error) {
	return nil, nil
}
