package app

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerJobExecutionResult struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_job_execution_result,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	RunnerJobExecutionID string `json:"runner_job_execution_id,omitzero" gorm:"defaultnull;notnull;index:idx_job_execution_result,unique" temporaljson:"runner_job_execution_id,omitzero,omitempty"`

	Success bool `json:"success,omitzero" temporaljson:"success,omitzero,omitempty"`

	ErrorCode     int           `json:"error_code,omitzero" temporaljson:"error_code,omitzero,omitempty"`
	ErrorMetadata pgtype.Hstore `json:"error_metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"error_metadata,omitzero,omitempty"`

	Contents        string `json:"contents,omitzero" gorm:"string" swaggertype:"string" temporaljson:"contents"`
	ContentsDisplay []byte `json:"contents_display,omitzero" gorm:"type:jsonb" swaggertype:"string" temporaljson:"-"`

	// columns for storage of gzipped contents and plans
	ContentsGzip        []byte `json:"contents_gzip,omitzero" gorm:"type:bytea" swaggertype:"string" temporaljson:"contents_binary"`
	ContentsDisplayGzip []byte `json:"contents_display_gzip,omitzero" gorm:"type:bytea" swaggertype:"string" temporaljson:"-"`
}

func (r *RunnerJobExecutionResult) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunnerJobExecutionResult{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (r *RunnerJobExecutionResult) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r *RunnerJobExecutionResult) GetContentsDisplayDecompressedBytes() ([]byte, error) {
	if len(r.ContentsDisplayGzip) == 0 {
		return []byte{}, nil
	}
	// take the []byte and uncompress it
	cdBuffer := bytes.NewReader(r.ContentsDisplayGzip)
	reader, err := gzip.NewReader(cdBuffer)
	if err != nil {
		return []byte{}, errors.Wrap(err, "unable to read contents display into gzip reader")
	}
	defer reader.Close()

	decompressedBytes, err := io.ReadAll(reader)
	if err != nil {
		return []byte{}, errors.Wrap(err, "unable to read contents display from gzip reader")
	}
	return decompressedBytes, nil
}

func (r *RunnerJobExecutionResult) GetContentsB64String() (string, error) {
	if len(r.ContentsGzip) == 0 {
		return "", nil
	}
	// base64 encode
	planB64 := base64.StdEncoding.EncodeToString(r.ContentsGzip) // NOTE(fd): internally we can use StdEncoding
	return planB64, nil

}

func (r *RunnerJobExecutionResult) GetContentsDisplayString() (string, error) {
	byts, err := r.GetContentsDisplayDecompressedBytes()
	if err != nil {
		return "", err
	}
	return string(byts), nil
}
