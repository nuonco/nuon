package activities

import (
	"context"
)

type SeedImportDBResponse struct {
	Size string `json:"size"`
}

type SeedImportDB struct {
	BackupFP       string `json:"backup_fp"`
	BackupS3Bucket string `json:"backup_s3_bucket"`
	BackupIAMRole  string `json:"backup_iam_role"`
}

// @temporal-gen activity
func (a *Activities) SeedImportDB(ctx context.Context, req SeedImportDB) (*SeedImportDBResponse, error) {
	return nil, nil
}
