package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppConfigSourceFile
// @Summary					get a single file from an app config's source archive
// @Description				Returns the raw contents of one file captured in the config's source archive. The path must exactly match a captured file.
// @Param					app_id		path	string	true	"app ID"
// @Param					config_id	path	string	true	"config ID"
// @Param					path		path	string	true	"file path relative to the config root"
// @Tags					apps
// @Produce					octet-stream
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{file}		file
// @Router					/v1/apps/{app_id}/configs/{config_id}/source-files/{path} [get]
func (s *service) GetAppConfigSourceFile(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	configID := ctx.Param("config_id")

	rawPath := strings.TrimPrefix(ctx.Param("path"), "/")
	filePath := filepath.ToSlash(filepath.Clean(rawPath))
	if filePath == "" || filePath == "." || filePath == ".." || strings.HasPrefix(filePath, "../") {
		ctx.Error(fmt.Errorf("source file not found: %w", gorm.ErrRecordNotFound))
		return
	}

	blobCtx := blobstore.WithBlobService(ctx.Request.Context(), s.blobSvc)
	archive, err := loadAppConfigSourceArchive(blobCtx, s.db, org.ID, appID, configID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if archive == nil {
		ctx.Error(fmt.Errorf("source file not found: %w", gorm.ErrRecordNotFound))
		return
	}

	contents, ok := archive.Files[filePath]
	if !ok {
		ctx.Error(fmt.Errorf("source file not found: %w", gorm.ErrRecordNotFound))
		return
	}

	checksum := sha256.Sum256([]byte(contents))
	etag := `"` + hex.EncodeToString(checksum[:]) + `"`
	if ctx.GetHeader("If-None-Match") == etag {
		ctx.Status(http.StatusNotModified)
		return
	}

	ctx.Header("ETag", etag)
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("Cache-Control", "private, max-age=31536000, immutable")
	if s.sourceFileSize != nil {
		s.sourceFileSize.Record(ctx.Request.Context(), int64(len(contents)))
	}
	ctx.Data(http.StatusOK, "application/octet-stream", []byte(contents))
}

func loadAppConfigSourceArchive(ctx context.Context, db *gorm.DB, orgID, appID, configID string) (*config.SourceArchive, error) {
	var appCfg app.AppConfig
	res := db.WithContext(ctx).
		Where(app.AppConfig{AppID: appID, OrgID: orgID}).
		First(&appCfg, "id = ?", configID)
	if res.Error != nil {
		return nil, res.Error
	}
	if appCfg.SourceConfig == nil {
		return nil, nil
	}
	raw, err := appCfg.SourceConfig.Get(ctx)
	if err != nil || raw == "" {
		return nil, err
	}
	var archive config.SourceArchive
	if err := json.Unmarshal([]byte(raw), &archive); err != nil {
		return nil, fmt.Errorf("unable to parse source config: %w", err)
	}
	return &archive, nil
}
