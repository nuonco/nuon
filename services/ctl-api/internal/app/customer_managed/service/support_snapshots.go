package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/runner/customer_managed/supportsnapshot"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	supportSnapshotIntegrityVerified   = "verified"
	supportSnapshotAssociationVerified = "verified"
)

type supportSnapshotResponse struct {
	ID                string                    `json:"id"`
	InstallID         string                    `json:"install_id"`
	CreatedAt         string                    `json:"created_at"`
	CapturedAt        string                    `json:"captured_at"`
	ArchiveSHA256     string                    `json:"archive_sha256"`
	ArchiveSize       int64                     `json:"archive_size"`
	SchemaVersion     int                       `json:"schema_version"`
	IntegrityStatus   string                    `json:"integrity_status"`
	AssociationStatus string                    `json:"association_status"`
	Manifest          supportsnapshot.Manifest  `json:"manifest"`
	Snapshot          *supportsnapshot.Snapshot `json:"snapshot,omitempty"`
}

func supportSnapshotResponseFromModel(snapshot app.InstallSupportSnapshot) supportSnapshotResponse {
	response := supportSnapshotResponse{
		ID: snapshot.ID, InstallID: snapshot.InstallID, CreatedAt: snapshot.CreatedAt.UTC().Format(time.RFC3339Nano),
		CapturedAt: snapshot.CapturedAt.UTC().Format(time.RFC3339Nano), ArchiveSHA256: snapshot.ArchiveSHA256,
		ArchiveSize: snapshot.ArchiveSize, SchemaVersion: snapshot.SchemaVersion,
		IntegrityStatus: snapshot.IntegrityStatus, AssociationStatus: snapshot.AssociationStatus, Manifest: snapshot.Manifest,
	}
	return response
}

func (s *service) supportSnapshotResponseWithData(ctx context.Context, record app.InstallSupportSnapshot) (supportSnapshotResponse, error) {
	response := supportSnapshotResponseFromModel(record)
	raw, err := record.SnapshotBlob.Get(blobstore.WithBlobService(ctx, s.blobSvc))
	if err != nil {
		return supportSnapshotResponse{}, fmt.Errorf("load support snapshot data: %w", err)
	}
	var snapshot supportsnapshot.Snapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return supportSnapshotResponse{}, fmt.Errorf("decode support snapshot data: %w", err)
	}
	response.Snapshot = &snapshot
	return response, nil
}

// @ID CreateInstallSupportSnapshot
// @Summary import a customer-managed install support snapshot
// @Tags installs
// @Accept application/octet-stream
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Param snapshot body string true "support snapshot archive"
// @Success 200 {object} supportSnapshotResponse
// @Success 201 {object} supportSnapshotResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} map[string]string
// @Failure 413 {object} map[string]string
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/support-snapshots [post]
func (s *service) CreateSupportSnapshot(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if !s.store.Configured() {
		ctx.JSON(http.StatusConflict, gin.H{"error": "customer-managed support snapshot storage is not configured"})
		return
	}
	install, registration, err := s.customerManagedInstall(ctx, org.ID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	tmp, err := os.CreateTemp("", "nuon-support-snapshot-*.tar.zst")
	if err != nil {
		ctx.Error(err)
		return
	}
	defer os.Remove(tmp.Name())
	hash := sha256.New()
	reader := http.MaxBytesReader(ctx.Writer, ctx.Request.Body, supportsnapshot.MaxArchiveSize)
	size, err := io.Copy(io.MultiWriter(tmp, hash), reader)
	if err != nil {
		tmp.Close()
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": fmt.Sprintf("support snapshot exceeds %d bytes", supportsnapshot.MaxArchiveSize)})
			return
		}
		ctx.Error(fmt.Errorf("read support snapshot: %w", err))
		return
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		tmp.Close()
		ctx.Error(err)
		return
	}
	archive, err := supportsnapshot.Read(tmp)
	if err != nil {
		tmp.Close()
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateSupportSnapshotAssociation(registration.Registration, archive.Snapshot); err != nil {
		tmp.Close()
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	var existing app.InstallSupportSnapshot
	lookup := app.InstallSupportSnapshot{OrgID: org.ID, ArchiveSHA256: digest}
	if err := s.db.WithContext(ctx).Where(lookup).First(&existing).Error; err == nil {
		tmp.Close()
		if err := s.syncBundleHistory(ctx, install, archive.Snapshot.BundleHistory); err != nil {
			ctx.Error(err)
			return
		}
		response, err := s.supportSnapshotResponseWithData(ctx, existing)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusOK, response)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tmp.Close()
		ctx.Error(err)
		return
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		tmp.Close()
		ctx.Error(err)
		return
	}
	replica, err := s.store.Publish(ctx, transport.PublishRequest{Body: tmp, Size: size, SHA256: digest})
	if closeErr := tmp.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	if err != nil {
		ctx.Error(fmt.Errorf("store support snapshot: %w", err))
		return
	}
	snapshotJSON, err := json.Marshal(archive.Snapshot)
	if err != nil {
		cleanupErr := s.store.Delete(ctx, replica)
		ctx.Error(errors.Join(fmt.Errorf("encode support snapshot data: %w", err), cleanupErr))
		return
	}
	snapshotBlob := &blobstore.Blob{}
	snapshotBlob.Set(string(snapshotJSON))
	snapshotBlob.SetContentType("application/json")
	record := app.InstallSupportSnapshot{
		OrgID: org.ID, InstallID: install.ID, ArchiveSHA256: digest, ArchiveSize: size,
		SchemaVersion: archive.Manifest.SchemaVersion, CapturedAt: archive.Manifest.CapturedAt,
		StorageProvider: replica.Provider, StorageRegion: replica.Region, StorageRef: replica.StorageRef, StorageVersion: replica.StorageVersion,
		IntegrityStatus: supportSnapshotIntegrityVerified, AssociationStatus: supportSnapshotAssociationVerified,
		Manifest: archive.Manifest, SnapshotBlob: snapshotBlob,
	}
	dbCtx := blobstore.WithBlobService(ctx, s.blobSvc)
	if err := s.db.WithContext(dbCtx).Omit("Org", "Install").Create(&record).Error; err != nil {
		cleanupErr := s.deleteSupportSnapshotObjects(ctx, replica, snapshotBlob)
		if cleanupErr != nil {
			ctx.Error(errors.Join(err, cleanupErr))
			return
		}
		if reloadErr := s.db.WithContext(ctx).Where(lookup).First(&existing).Error; reloadErr == nil {
			response, responseErr := s.supportSnapshotResponseWithData(ctx, existing)
			if responseErr != nil {
				ctx.Error(responseErr)
				return
			}
			ctx.JSON(http.StatusOK, response)
			return
		}
		ctx.Error(err)
		return
	}
	if err := s.syncBundleHistory(ctx, install, archive.Snapshot.BundleHistory); err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.supportSnapshotResponseWithData(ctx, record)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

func (s *service) deleteSupportSnapshotObjects(ctx context.Context, replica transport.Replica, snapshotBlob *blobstore.Blob) error {
	var errs []error
	if err := s.store.Delete(ctx, replica); err != nil {
		errs = append(errs, fmt.Errorf("delete support snapshot archive: %w", err))
	}
	if snapshotBlob != nil && snapshotBlob.Metadata().S3Key != "" {
		if err := s.blobSvc.Delete(ctx, snapshotBlob.Metadata().S3Key); err != nil {
			errs = append(errs, fmt.Errorf("delete support snapshot data: %w", err))
		}
	}
	return errors.Join(errs...)
}

// @ID ListInstallSupportSnapshots
// @Summary list imported support snapshots for a customer-managed install
// @Tags installs
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Success 200 {array} supportSnapshotResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/support-snapshots [get]
func (s *service) ListSupportSnapshots(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	install, _, err := s.customerManagedInstall(ctx, org.ID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	var records []app.InstallSupportSnapshot
	if err := s.db.WithContext(ctx).Where(app.InstallSupportSnapshot{OrgID: org.ID, InstallID: install.ID}).Order("captured_at DESC").Find(&records).Error; err != nil {
		ctx.Error(err)
		return
	}
	response := make([]supportSnapshotResponse, 0, len(records))
	for _, record := range records {
		response = append(response, supportSnapshotResponseFromModel(record))
	}
	ctx.JSON(http.StatusOK, response)
}

// @ID GetInstallSupportSnapshot
// @Summary get one imported customer-managed install support snapshot
// @Tags installs
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Param snapshot_id path string true "Support snapshot ID"
// @Success 200 {object} supportSnapshotResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/support-snapshots/{snapshot_id} [get]
func (s *service) GetSupportSnapshot(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	install, _, err := s.customerManagedInstall(ctx, org.ID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	var record app.InstallSupportSnapshot
	query := app.InstallSupportSnapshot{ID: ctx.Param("snapshot_id"), OrgID: org.ID, InstallID: install.ID}
	if err := s.db.WithContext(ctx).Where(query).First(&record).Error; err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.supportSnapshotResponseWithData(ctx, record)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *service) customerManagedInstall(ctx *gin.Context, orgID, installID string) (*app.Install, app.InstallRegistration, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{ID: installID, OrgID: orgID}).First(&install).Error; err != nil {
		return nil, app.InstallRegistration{}, err
	}
	var registration app.InstallRegistration
	if err := s.db.WithContext(ctx).Where(app.InstallRegistration{InstallID: install.ID, OrgID: orgID}).Order("imported_at DESC").First(&registration).Error; err != nil {
		return nil, app.InstallRegistration{}, fmt.Errorf("install %s is not customer-managed: %w", installID, err)
	}
	return &install, registration, nil
}

func validateSupportSnapshotAssociation(installed customermanaged.InstallationRegistration, snapshot supportsnapshot.Snapshot) error {
	registration := snapshot.Registration
	if registration.RegistrationID != installed.RegistrationID {
		return errors.New("support snapshot registration does not match this install")
	}
	if registration.InstallID != installed.InstallID {
		return errors.New("support snapshot installation identity does not match this install")
	}
	if registration.ReleaseID != installed.ReleaseID || registration.PackageID != installed.PackageID ||
		!strings.EqualFold(registration.ReleaseDigest, installed.ReleaseDigest) || !strings.EqualFold(registration.PackageDigest, installed.PackageDigest) ||
		!strings.EqualFold(registration.BundleDigest, installed.BundleDigest) || !strings.EqualFold(registration.ArchiveDigest, installed.ArchiveDigest) {
		return errors.New("support snapshot bundle identity does not match this install")
	}
	return nil
}
