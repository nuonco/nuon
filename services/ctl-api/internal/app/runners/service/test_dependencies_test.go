package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	githubclient "github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/heartbeater"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

func resetQueueSignals(t testing.TB, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&app.QueueSignal{}).Error)
}

func seedRunnerSignalsQueue(t testing.TB, db *gorm.DB, runnerID string) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&app.Queue{}).Where(&app.Queue{OwnerID: runnerID, OwnerType: "runners", Name: "runner-signals"}).Count(&count).Error)
	if count != 0 {
		return
	}
	var account app.Account
	require.NoError(t, db.First(&account).Error)
	require.NoError(t, db.Create(&app.Queue{CreatedByID: account.ID, OwnerID: runnerID, OwnerType: "runners", Name: "runner-signals", MaxInFlight: 10, MaxDepth: 50}).Error)
}

type testDBParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	DB        *gorm.DB `name:"psql"`
	CHDB      *gorm.DB `name:"ch"`
}

func testDependencyOptions() fx.Option {
	return fx.Options(fx.Replace(githubclient.NewClient(nil)), fx.Decorate(func(blobstore.Service) blobstore.Service {
		return &memBlobStore{}
	}), fx.Decorate(func(cfg *internal.Config) *internal.Config {
		cfg.HeartbeaterFlushInterval = 50 * time.Millisecond
		return cfg
	}), fx.Provide(
		NewRunnerHeartbeatCache,
		NewRunnerJobWakeRegistry,
		heartbeater.New,
		kafka.New,
	), fx.Invoke(func(params testDBParams) error {
		sqlDB, err := params.DB.DB()
		if err != nil {
			return err
		}
		sqlDB.SetMaxOpenConns(2)
		sqlDB.SetMaxIdleConns(1)
		chSQLDB, err := params.CHDB.DB()
		if err != nil {
			return err
		}
		if err := params.DB.Callback().Create().After("gorm:create").Register("runner_service_test:seed_signals_queue", func(tx *gorm.DB) {
			runner, ok := tx.Statement.Dest.(*app.Runner)
			if !ok {
				return
			}
			tx.AddError(tx.Session(&gorm.Session{NewDB: true}).Create(&app.Queue{
				OwnerID:     runner.ID,
				OwnerType:   "runners",
				Name:        "runner-signals",
				MaxInFlight: 10,
				MaxDepth:    50,
			}).Error)
		}); err != nil {
			return err
		}
		if err := params.DB.Callback().Create().Before("gorm:create").Register("runner_service_test:seed_signal_creator", func(tx *gorm.DB) {
			signal, ok := tx.Statement.Dest.(*app.QueueSignal)
			if !ok || signal.CreatedByID != "" {
				return
			}
			var queue app.Queue
			if err := tx.Session(&gorm.Session{NewDB: true}).First(&queue, app.Queue{ID: signal.QueueID}).Error; err != nil {
				tx.AddError(err)
				return
			}
			signal.CreatedByID = queue.CreatedByID
		}); err != nil {
			return err
		}
		params.Lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
			if err := sqlDB.Close(); err != nil {
				return err
			}
			return chSQLDB.Close()
		}})
		return params.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&app.QueueSignal{}).Error
	}))
}

func withOrgAccess(acc *app.Account, org *app.Org) *app.Account {
	acc.AllPermissions = permissions.Set{org.ID: permissions.PermissionAll}
	return acc
}

type memBlobStore struct {
	blobs sync.Map
}

func (m *memBlobStore) Upload(_ context.Context, key string, data []byte) error {
	m.blobs.Store(key, bytes.Clone(data))
	return nil
}

func (m *memBlobStore) Download(_ context.Context, key string) ([]byte, error) {
	v, ok := m.blobs.Load(key)
	if !ok {
		return nil, fmt.Errorf("blob %s not found", key)
	}
	return bytes.Clone(v.([]byte)), nil
}

func (m *memBlobStore) UploadStream(ctx context.Context, key string, reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), m.Upload(ctx, key, data)
}

func (m *memBlobStore) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	data, err := m.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *memBlobStore) GetMetadata(ctx context.Context, key string) (int64, string, error) {
	data, err := m.Download(ctx, key)
	if err != nil {
		return 0, "", err
	}
	return int64(len(data)), "application/octet-stream", nil
}
