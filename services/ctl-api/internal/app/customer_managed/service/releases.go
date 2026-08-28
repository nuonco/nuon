package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	pkgconfig "github.com/nuonco/nuon/pkg/config"
	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	customermanagedapp "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type createReleaseRequest struct {
	AppConfigID string                            `json:"app_config_id" binding:"required"`
	Runbooks    []customermanaged.RunbookTemplate `json:"runbooks,omitempty"`
}

// @ID CreateAppRelease
// @Summary create an immutable application release
// @Tags releases
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param request body createReleaseRequest true "release request"
// @Success 200 {object} app.AppRelease
// @Success 201 {object} app.AppRelease
// @Failure 400 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 412 {object} map[string]string
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/releases [post]
func (s *service) CreateRelease(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	var req createReleaseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	release, created, err := s.createRelease(ctx, org.ID, ctx.Param("app_id"), req)
	if err != nil {
		if precondition, ok := err.(preconditionError); ok {
			ctx.JSON(http.StatusPreconditionFailed, gin.H{"error": precondition.msg})
			return
		}
		ctx.Error(fmt.Errorf("unable to create app release: %w", err))
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	ctx.JSON(status, release)
}

func (s *service) createRelease(ctx context.Context, orgID, appID string, req createReleaseRequest) (*app.AppRelease, bool, error) {
	cfg, err := s.appsHelpers.GetCustomerManagedAppConfig(ctx, orgID, appID, req.AppConfigID)
	if err != nil {
		return nil, false, err
	}
	selection, err := s.resolveActiveBuilds(ctx, orgID, appID, cfg)
	if err != nil {
		return nil, false, err
	}
	runbooks, _, err := canonicalBundleRunbooks(req.Runbooks)
	if err != nil {
		return nil, false, fmt.Errorf("canonicalize release runbooks: %w", err)
	}
	members, err := s.releaseMembers(ctx, orgID, cfg, selection, runbooks)
	if err != nil {
		return nil, false, err
	}
	sourceArchive, err := s.authoredSourceArchive(ctx, cfg)
	if err != nil {
		return nil, false, err
	}
	members, err = appendAuthoredReleaseMembers(members, sourceArchive)
	if err != nil {
		return nil, false, err
	}
	runtime, err := releaseRuntime(cfg)
	if err != nil {
		return nil, false, preconditionError{msg: err.Error()}
	}
	runtimeDigest := customermanagedapp.ObjectDigest(runtime)
	semanticDigest := semanticReleaseDigest(cfg.ID, runtimeDigest, members)
	var existing app.AppRelease
	err = s.db.WithContext(ctx).
		Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Where(app.AppReleaseMember{OrgID: orgID}) }).
		Preload("Packages", func(db *gorm.DB) *gorm.DB { return db.Where(app.ReleasePackage{OrgID: orgID}) }).
		Where(app.AppRelease{OrgID: orgID, AppID: appID, SemanticDigest: semanticDigest}).First(&existing).Error
	if err == nil {
		if err := s.hydrateReleaseDefinitions(ctx, &existing); err != nil {
			return nil, false, err
		}
		return &existing, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}
	release := app.AppRelease{
		OrgID: orgID, AppID: appID, AppConfigID: cfg.ID,
		SandboxBuildID: selection.sandboxBuildID, ComponentBuildIDs: selection.componentBuildIDs,
		Runbooks: runbooks, Runtime: runtime, RuntimeDigest: runtimeDigest,
		SchemaVersion: bundle.CurrentSchemaVersion, SemanticDigest: semanticDigest,
		Status: app.AppReleaseStatusReady, StatusDescription: "release ready",
	}
	authoredDefinitions := authoredReleaseDefinitions(sourceArchive)
	definitionsBlob, err := newReleaseDefinitionsBlob(members, authoredDefinitions, sourceArchive)
	if err != nil {
		return nil, false, err
	}
	release.DefinitionsBlob = definitionsBlob
	dbCtx := blobstore.WithBlobService(ctx, s.blobSvc)
	err = s.db.WithContext(dbCtx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&release).Error; err != nil {
			return err
		}
		for i := range members {
			members[i].ReleaseID = release.ID
			members[i].OrgID = orgID
		}
		return tx.Create(&members).Error
	})
	if err != nil {
		if key := definitionsBlob.Metadata().S3Key; key != "" {
			if cleanupErr := s.blobSvc.Delete(ctx, key); cleanupErr != nil {
				return nil, false, errors.Wrapf(err, "cleanup release definitions blob: %v", cleanupErr)
			}
		}
		return nil, false, err
	}
	release.Members = members
	if sourceArchive != nil {
		release.SourceFiles = (customermanaged.ReleaseArchive{Files: sourceArchive.Files}).FileList()
	}
	return &release, true, nil
}

func newReleaseDefinitionsBlob(members []app.AppReleaseMember, authored map[string]string, sourceArchive *pkgconfig.SourceArchive) (*blobstore.Blob, error) {
	definitions := customermanaged.ReleaseArchive{SchemaVersion: 2, Members: make(map[string]string, len(members))}
	if sourceArchive != nil {
		definitions.Files = sourceArchive.Files
	}
	for _, member := range members {
		key := releaseMemberKey(member)
		definitions.Members[key] = member.ConfigTOML
		if source, ok := authored[key]; ok {
			definitions.Members[key] = source
		}
	}
	raw, err := json.Marshal(definitions)
	if err != nil {
		return nil, fmt.Errorf("encode release definitions: %w", err)
	}
	blob := &blobstore.Blob{}
	blob.Set(string(raw))
	blob.SetContentType("application/json")
	return blob, nil
}

func (s *service) authoredSourceArchive(ctx context.Context, cfg *app.AppConfig) (*pkgconfig.SourceArchive, error) {
	if cfg.SourceConfig == nil || !cfg.SourceConfig.IsSet() {
		return nil, nil
	}
	raw, err := cfg.SourceConfig.Get(blobstore.WithBlobService(ctx, s.blobSvc))
	if err != nil {
		return nil, fmt.Errorf("load authored app config: %w", err)
	}
	var archive pkgconfig.SourceArchive
	if err := json.Unmarshal([]byte(raw), &archive); err != nil {
		return nil, fmt.Errorf("decode authored app config: %w", err)
	}
	if err := archive.ReindexMembers(); err != nil {
		return nil, fmt.Errorf("index authored app config: %w", err)
	}
	return &archive, nil
}

func authoredReleaseDefinitions(archive *pkgconfig.SourceArchive) map[string]string {
	if archive == nil {
		return nil
	}
	definitions := make(map[string]string, len(archive.Members))
	for key := range archive.Members {
		if source, ok := archive.MemberSource(key); ok {
			definitions[key] = source
		}
	}
	return definitions
}

func appendAuthoredReleaseMembers(members []app.AppReleaseMember, archive *pkgconfig.SourceArchive) ([]app.AppReleaseMember, error) {
	if archive == nil {
		return members, nil
	}
	existing := make(map[string]struct{}, len(members))
	memberPaths := make(map[string]struct{}, len(archive.Members))
	for _, path := range archive.Members {
		memberPaths[path] = struct{}{}
	}
	for _, member := range members {
		existing[releaseMemberKey(member)] = struct{}{}
	}
	for key, path := range archive.Members {
		if _, ok := existing[key]; ok {
			continue
		}
		source, ok := archive.MemberSource(key)
		if !ok {
			continue
		}
		var definition map[string]any
		if err := toml.Unmarshal([]byte(source), &definition); err != nil {
			return nil, fmt.Errorf("decode authored release member %s: %w", path, err)
		}
		parts := strings.SplitN(key, ":", 2)
		digest := customermanagedapp.ObjectDigest(definition)
		members = append(members, app.AppReleaseMember{
			Kind: parts[0], LogicalName: parts[1], ConfigDigest: digest, ConfigTOML: source,
			ContentDigest: digest, SourceType: "toml", SourceIdentity: map[string]any{"path": path},
		})
	}
	for path, contents := range archive.Files {
		if _, isDefinition := memberPaths[path]; isDefinition {
			continue
		}
		contentDigest := digest.FromString(contents).String()
		members = append(members, app.AppReleaseMember{
			Kind: "source_file", LogicalName: path, ConfigDigest: contentDigest,
			ContentDigest: contentDigest, SourceType: strings.TrimPrefix(filepath.Ext(path), "."), SourceIdentity: map[string]any{"path": path},
		})
	}
	sort.Slice(members, func(i, j int) bool {
		if members[i].Kind == members[j].Kind {
			return members[i].LogicalName < members[j].LogicalName
		}
		return members[i].Kind < members[j].Kind
	})
	return members, nil
}

func (s *service) hydrateReleaseDefinitions(ctx context.Context, release *app.AppRelease) error {
	if release.DefinitionsBlob == nil || !release.DefinitionsBlob.IsSet() {
		return nil
	}
	raw, err := release.DefinitionsBlob.Get(blobstore.WithBlobService(ctx, s.blobSvc))
	if err != nil {
		return fmt.Errorf("load release definitions: %w", err)
	}
	var definitions customermanaged.ReleaseArchive
	if err := json.Unmarshal([]byte(raw), &definitions); err != nil {
		return fmt.Errorf("decode release definitions: %w", err)
	}
	for i := range release.Members {
		release.Members[i].ConfigTOML = definitions.Members[releaseMemberKey(release.Members[i])]
	}
	release.SourceFiles = definitions.FileList()
	return nil
}

func releaseMemberKey(member app.AppReleaseMember) string {
	return member.Kind + ":" + member.LogicalName
}

func (s *service) releaseMembers(ctx context.Context, orgID string, cfg *app.AppConfig, selection bundleBuildSelection, runbooks []customermanaged.RunbookTemplate) ([]app.AppReleaseMember, error) {
	members := make([]app.AppReleaseMember, 0, len(cfg.ComponentConfigConnections)+len(cfg.ActionWorkflowConfigs)+len(runbooks)+1)
	for _, connection := range cfg.ComponentConfigConnections {
		definition, err := customermanagedapp.CanonicalComponentDefinition(connection, cfg.ComponentConfigConnections)
		if err != nil {
			return nil, fmt.Errorf("canonicalize component %s: %w", connection.ComponentName, err)
		}
		buildID := selection.componentBuildIDs[connection.ID]
		var build app.ComponentBuild
		if err := s.db.WithContext(ctx).Where(app.ComponentBuild{ID: buildID, OrgID: orgID, ComponentConfigConnectionID: connection.ID}).First(&build).Error; err != nil {
			return nil, err
		}
		name := connection.ComponentName
		if name == "" {
			name = connection.ComponentID
		}
		configTOML, err := canonicalDefinitionTOML(definition)
		if err != nil {
			return nil, fmt.Errorf("encode component %s definition as TOML: %w", name, err)
		}
		members = append(members, app.AppReleaseMember{
			Kind: "component", LogicalName: name, ComponentConfigConnectionID: connection.ID,
			ComponentID: connection.ComponentID, BuildID: build.ID, ConfigDigest: customermanagedapp.ObjectDigest(definition), ConfigTOML: configTOML,
			ContentDigest: customermanagedapp.ObjectDigest(map[string]string{"build_id": build.ID, "source_digest": build.SourceDigest, "source_checksum": build.SourceChecksum}),
			SourceType:    string(connection.Type), SourceIdentity: map[string]any{"source_ref": build.SourceRef, "source_digest": build.SourceDigest},
		})
	}
	sandboxDefinition, err := customermanagedapp.CanonicalObject(cfg.SandboxConfig)
	if err != nil {
		return nil, fmt.Errorf("canonicalize sandbox: %w", err)
	}
	sandboxTOML, err := canonicalDefinitionTOML(sandboxDefinition)
	if err != nil {
		return nil, fmt.Errorf("encode sandbox definition as TOML: %w", err)
	}
	members = append(members, app.AppReleaseMember{
		Kind: "sandbox", LogicalName: "sandbox", AppSandboxConfigID: cfg.SandboxConfig.ID,
		BuildID: selection.sandboxBuildID, ConfigDigest: customermanagedapp.ObjectDigest(sandboxDefinition), ConfigTOML: sandboxTOML,
		ContentDigest: customermanagedapp.ObjectDigest(map[string]string{"build_id": selection.sandboxBuildID}), SourceType: cfg.SandboxConfig.Type,
	})
	for _, actionConfig := range cfg.ActionWorkflowConfigs {
		name := actionConfig.ActionWorkflow.Name
		if name == "" {
			name = actionConfig.ActionWorkflowID
		}
		definition := customermanagedapp.CanonicalActionDefinition(actionConfig, cfg.ComponentConfigConnections)
		digest := customermanagedapp.ObjectDigest(definition)
		configTOML, err := canonicalDefinitionTOML(definition)
		if err != nil {
			return nil, fmt.Errorf("encode action %s definition as TOML: %w", name, err)
		}
		members = append(members, app.AppReleaseMember{Kind: "action", LogicalName: name, ActionWorkflowID: actionConfig.ActionWorkflowID, ConfigDigest: digest, ConfigTOML: configTOML, ContentDigest: digest})
	}
	runbookNames := make(map[string]struct{})
	var runbookConfigs []app.RunbookConfig
	if err := s.db.WithContext(ctx).Preload("Runbook").Preload("Steps").Preload("Inputs").Where(app.RunbookConfig{OrgID: orgID, AppConfigID: cfg.ID}).Find(&runbookConfigs).Error; err != nil {
		return nil, err
	}
	for _, runbookConfig := range runbookConfigs {
		name := runbookConfig.Runbook.Name
		if name == "" {
			name = runbookConfig.RunbookID
		}
		definition := customermanagedapp.CanonicalRunbookDefinition(runbookConfig, cfg.ActionWorkflowConfigs)
		digest := customermanagedapp.ObjectDigest(definition)
		configTOML, err := canonicalDefinitionTOML(definition)
		if err != nil {
			return nil, fmt.Errorf("encode runbook %s definition as TOML: %w", name, err)
		}
		members = append(members, app.AppReleaseMember{Kind: "runbook", LogicalName: name, ConfigDigest: digest, ConfigTOML: configTOML, ContentDigest: digest, SourceIdentity: map[string]any{"runbook_id": runbookConfig.RunbookID}})
		runbookNames[name] = struct{}{}
	}
	for _, runbook := range runbooks {
		if _, exists := runbookNames[runbook.Name]; exists {
			return nil, fmt.Errorf("runbook %q is defined by both the app config and release request", runbook.Name)
		}
		digest := customermanagedapp.ObjectDigest(runbook)
		configTOML, err := canonicalDefinitionTOML(runbook)
		if err != nil {
			return nil, fmt.Errorf("encode runbook %s definition as TOML: %w", runbook.Name, err)
		}
		members = append(members, app.AppReleaseMember{Kind: "runbook", LogicalName: runbook.Name, ConfigDigest: digest, ConfigTOML: configTOML, ContentDigest: digest})
	}
	sort.Slice(members, func(i, j int) bool {
		if members[i].Kind == members[j].Kind {
			return members[i].LogicalName < members[j].LogicalName
		}
		return members[i].Kind < members[j].Kind
	})
	return members, nil
}

func canonicalDefinitionTOML(definition any) (string, error) {
	raw, err := pkgconfig.ToTOML(definition)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func releaseRuntime(cfg *app.AppConfig) (app.AppReleaseRuntime, error) {
	if cfg.CustomerManagedRuntime == nil {
		return app.AppReleaseRuntime{}, fmt.Errorf("app config must define customer_managed runtime artifacts")
	}
	runtime := *cfg.CustomerManagedRuntime
	if runtime.RunnerImageURL == "" || runtime.RunnerImageTag == "" {
		return app.AppReleaseRuntime{}, fmt.Errorf("customer_managed runner_image_url and runner_image_tag are required")
	}
	platform, ok := runtime.Platforms["linux/amd64"]
	if !ok {
		return app.AppReleaseRuntime{}, fmt.Errorf("customer_managed platform linux/amd64 is required")
	}
	if platform.PortalBinaryURL == "" || platform.RunnerBinaryURL == "" {
		return app.AppReleaseRuntime{}, fmt.Errorf("customer_managed platform linux/amd64 requires portal_binary_url and runner_binary_url")
	}
	return runtime, nil
}

func semanticReleaseDigest(appConfigID, runtimeDigest string, members []app.AppReleaseMember) string {
	members = append([]app.AppReleaseMember(nil), members...)
	sort.Slice(members, func(i, j int) bool {
		if members[i].Kind == members[j].Kind {
			return members[i].LogicalName < members[j].LogicalName
		}
		return members[i].Kind < members[j].Kind
	})
	identity := make([]map[string]string, 0, len(members))
	for _, member := range members {
		identity = append(identity, map[string]string{
			"kind": member.Kind, "name": member.LogicalName, "build_id": member.BuildID,
			"config_digest": member.ConfigDigest, "content_digest": member.ContentDigest,
		})
	}
	return customermanagedapp.ObjectDigest(map[string]any{"schema_version": bundle.CurrentSchemaVersion, "app_config_id": appConfigID, "runtime_digest": runtimeDigest, "members": identity})
}

// @ID ListAppReleases
// @Summary list immutable application releases
// @Tags releases
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param offset query int false "offset of results to return" Default(0)
// @Param limit query int false "maximum number of results to return" Default(10)
// @Success 200 {array} app.AppRelease
// @Failure 403 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/releases [get]
func (s *service) ListReleases(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var releases []app.AppRelease
	result := s.db.WithContext(ctx).Scopes(scopes.WithOffsetPagination).
		Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Where(app.AppReleaseMember{OrgID: org.ID}) }).
		Preload("Packages", func(db *gorm.DB) *gorm.DB { return db.Where(app.ReleasePackage{OrgID: org.ID}) }).
		Where(app.AppRelease{OrgID: org.ID, AppID: ctx.Param("app_id")}).Order("created_at DESC").Find(&releases)
	if result.Error != nil {
		ctx.Error(result.Error)
		return
	}
	releases, err = db.HandlePaginatedResponse(ctx, releases)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, releases)
}

// @ID GetAppRelease
// @Summary get an immutable application release
// @Tags releases
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param release_id path string true "release ID"
// @Success 200 {object} app.AppRelease
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/releases/{release_id} [get]
func (s *service) GetRelease(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	release, err := s.getRelease(ctx, org.ID, ctx.Param("app_id"), ctx.Param("release_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, release)
}

func (s *service) getRelease(ctx context.Context, orgID, appID, releaseID string) (*app.AppRelease, error) {
	var release app.AppRelease
	if err := s.db.WithContext(ctx).
		Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Where(app.AppReleaseMember{OrgID: orgID}) }).
		Preload("Packages", func(db *gorm.DB) *gorm.DB { return db.Where(app.ReleasePackage{OrgID: orgID}) }).
		Where(app.AppRelease{ID: releaseID, OrgID: orgID, AppID: appID}).First(&release).Error; err != nil {
		return nil, err
	}
	if err := s.hydrateReleaseDefinitions(ctx, &release); err != nil {
		return nil, err
	}
	return &release, nil
}

type releaseFileContentResponse struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	MediaType string `json:"media_type"`
}

// @ID GetAppReleaseFileContent
// @Summary get one authored file from an immutable application release
// @Tags releases
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "app ID"
// @Param release_id path string true "release ID"
// @Param path query string true "release-relative file path"
// @Success 200 {object} releaseFileContentResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Router /v1/apps/{app_id}/releases/{release_id}/files/content [get]
func (s *service) GetReleaseFileContent(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	content, err := s.getReleaseFileContent(ctx, org.ID, ctx.Param("app_id"), ctx.Param("release_id"), ctx.Query("path"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, content)
}

func (s *service) getReleaseFileContent(ctx context.Context, orgID, appID, releaseID, path string) (*releaseFileContentResponse, error) {
	var release app.AppRelease
	if err := s.db.WithContext(ctx).
		Where(app.AppRelease{ID: releaseID, OrgID: orgID, AppID: appID}).
		First(&release).Error; err != nil {
		return nil, err
	}
	if release.DefinitionsBlob == nil || !release.DefinitionsBlob.IsSet() {
		return nil, gorm.ErrRecordNotFound
	}
	raw, err := release.DefinitionsBlob.Get(blobstore.WithBlobService(ctx, s.blobSvc))
	if err != nil {
		return nil, fmt.Errorf("load release files: %w", err)
	}
	var archive customermanaged.ReleaseArchive
	if err := json.Unmarshal([]byte(raw), &archive); err != nil {
		return nil, fmt.Errorf("decode release files: %w", err)
	}
	content, ok := archive.Files[path]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	for _, file := range archive.FileList() {
		if file.Path == path {
			return &releaseFileContentResponse{
				Path: file.Path, Content: content, Digest: file.Digest, Size: file.Size, MediaType: file.MediaType,
			}, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *service) customerManagedInstallsEnabled(ctx *gin.Context) bool {
	if s.features == nil {
		ctx.Error(fmt.Errorf("customer-managed installs feature client is unavailable"))
		return false
	}
	enabled, err := s.features.FeatureEnabled(ctx, app.OrgFeatureCustomerManagedInstalls)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check customer-managed installs feature: %w", err))
		return false
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureCustomerManagedInstalls))
		return false
	}
	return true
}
