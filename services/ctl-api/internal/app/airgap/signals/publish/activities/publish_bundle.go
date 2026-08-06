package activities

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gorm.io/gorm"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	runnerairgap "github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/pkg/runner/oci"
	"github.com/nuonco/nuon/pkg/runner/oci/bundle"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/airgap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/airgap/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks/cloudformation"
)

const (
	maxBundleContent = int64(20 << 30)
	maxBundleBlob    = int64(5 << 30)
	maxFetchedAsset  = int64(100 << 20)

	stackAssetYAMLMediaType  = "application/yaml"
	stackAssetShellMediaType = "text/x-shellscript"

	// Mirrors DefaultAWSRunnerInitScript in installs/signals/generateinstallstackversion;
	// that package pulls in the temporal worker stack, so the URL is duplicated here.
	defaultAWSRunnerInitScriptURL = "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh#default"
)

var assetHTTPClient = &http.Client{
	Timeout: 2 * time.Minute,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type PublishBundleRequest struct {
	BundleID string                         `validate:"required"`
	Runbooks []runnerairgap.RunbookTemplate `json:"runbooks,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 180m
func (a *Activities) PublishBundle(ctx context.Context, req *PublishBundleRequest) error {
	var published app.AirgapBundle
	if err := a.db.WithContext(ctx).Where(app.AirgapBundle{ID: req.BundleID}).First(&published).Error; err != nil {
		return fmt.Errorf("load bundle: %w", err)
	}
	if published.ManifestDigest != "" && published.OCIRootDigest != "" && published.TransportChecksum != "" {
		var replicas []app.AirgapBundleTransportReplica
		if err := a.db.WithContext(ctx).Where(app.AirgapBundleTransportReplica{BundleID: published.ID, OrgID: published.OrgID}).Find(&replicas).Error; err != nil {
			return fmt.Errorf("load bundle replicas: %w", err)
		}
		for _, replica := range replicas {
			if replica.VerifiedAt != nil {
				return a.db.WithContext(ctx).Model(&published).Updates(app.AirgapBundle{Status: app.AirgapBundleStatusActive, StatusDescription: "bundle published and verified"}).Error
			}
		}
	}
	cfg, err := a.appsHelpers.GetAirgapAppConfig(ctx, published.OrgID, published.AppID, published.AppConfigID)
	if err != nil {
		return err
	}
	report := airgap.Qualify(cfg, published.TargetPlatform)
	if !report.Qualified {
		return fmt.Errorf("app config does not qualify for air-gap export")
	}
	envelope, err := airgap.CompilePlanEnvelope(ctx, a.db, a.v, published.OrgID, published.AppID, cfg, published.SandboxBuildID, published.ComponentBuildIDs, req.Runbooks, &report)
	if err != nil {
		return fmt.Errorf("compile plan envelope: %w", err)
	}
	logical, roots, artifacts, provenance, err := a.bundleInputs(ctx, &published, cfg, "")
	if err != nil {
		return err
	}
	pins, err := a.bundlePins(ctx, &published, cfg)
	if err != nil {
		return err
	}
	if err := airgap.RewriteEnvelopeForBundle(ctx, a.db, envelope, pins); err != nil {
		return fmt.Errorf("rewrite plan envelope to pinned bundle members: %w", err)
	}
	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal plan envelope: %w", err)
	}
	reportJSON, _ := json.Marshal(report)
	provenanceJSON, _ := json.Marshal(provenance)
	f, err := os.CreateTemp("", "nuon-airgap-*.tar.zst")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	result, err := bundle.GenerateWithOptions(ctx, f, logical, bundle.Documents{Provenance: provenanceJSON, QualificationReport: reportJSON, PlanEnvelope: envelopeJSON}, roots, bundle.GenerateOptions{MaxContentBytes: maxBundleContent, MaxBlobBytes: maxBundleBlob})
	if err != nil {
		return fmt.Errorf("generate bundle: %w", err)
	}
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	replica, err := a.store.Publish(ctx, transport.PublishRequest{Body: f, Size: stat.Size(), SHA256: result.TransportSHA256})
	if err != nil {
		return fmt.Errorf("publish bundle: %w", err)
	}
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := app.AirgapBundle{ManifestDigest: result.ManifestDescriptor.Digest.String(), OCIRootDigest: result.BundleDescriptor.Digest.String(), TransportChecksum: result.TransportSHA256, Size: stat.Size(), Status: app.AirgapBundleStatusActive, StatusDescription: "bundle published and verified"}
		if err := tx.Model(&app.AirgapBundle{ID: published.ID}).Updates(updates).Error; err != nil {
			return err
		}
		for i := range artifacts {
			artifacts[i].BundleID = published.ID
			artifacts[i].OrgID = published.OrgID
		}
		if err := tx.Where(app.AirgapBundleArtifact{BundleID: published.ID}).Delete(&app.AirgapBundleArtifact{}).Error; err != nil {
			return err
		}
		if len(artifacts) > 0 {
			if err := tx.Create(&artifacts).Error; err != nil {
				return err
			}
		}
		verified := replica.VerifiedAt
		r := app.AirgapBundleTransportReplica{BundleID: published.ID, OrgID: published.OrgID, Provider: replica.Provider, Region: replica.Region, StorageRef: replica.StorageRef, StorageVersion: replica.StorageVersion, TransportChecksum: replica.TransportChecksum, Size: replica.Size, VerifiedAt: &verified}
		return tx.Create(&r).Error
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *Activities) bundlePins(ctx context.Context, published *app.AirgapBundle, cfg *app.AppConfig) (airgap.BundlePins, error) {
	pins := airgap.BundlePins{SandboxBuildID: published.SandboxBuildID, ComponentBuildIDs: map[string]string{}}
	for _, connection := range cfg.ComponentConfigConnections {
		buildID := published.ComponentBuildIDs[connection.ID]
		if buildID == "" {
			return pins, fmt.Errorf("bundle has no pinned build for component %s", connection.ComponentName)
		}
		pins.ComponentBuildIDs[connection.ComponentID] = buildID
	}
	var currentApp app.App
	if err := a.db.WithContext(ctx).Preload("Repository").Where(app.App{ID: published.AppID, OrgID: published.OrgID}).First(&currentApp).Error; err != nil {
		return pins, err
	}
	// No auth block on purpose: the airgap runner serves this source from the
	// bundle by tag, so the registry is a fail-closed placeholder.
	pins.SandboxRegistry = &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypeECR, Repository: currentApp.Repository.RepositoryURI, Region: currentApp.Repository.Region}
	return pins, nil
}

func (a *Activities) bundleInputs(ctx context.Context, published *app.AirgapBundle, cfg *app.AppConfig, referenceInstallID string) (bundle.LogicalManifest, []bundle.Root, []app.AirgapBundleArtifact, map[string]any, error) {
	logical := bundle.LogicalManifest{SchemaVersion: 1, Target: bundle.Target{OS: "linux", Architecture: "amd64"}}
	var currentApp app.App
	if err := a.db.WithContext(ctx).Preload("Repository").Where(app.App{ID: published.AppID, OrgID: published.OrgID}).First(&currentApp).Error; err != nil {
		return logical, nil, nil, nil, err
	}
	repoCfg := &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypeECR, Repository: currentApp.Repository.RepositoryURI, Region: currentApp.Repository.Region, ECRAuth: &credentials.Config{Region: currentApp.Repository.Region, AssumeRole: &credentials.AssumeRoleConfig{RoleARN: a.cfg.ManagementIAMRoleARN, SessionName: "airgap-bundle"}}}
	repo, err := oci.GetRepo(ctx, repoCfg)
	if err != nil {
		return logical, nil, nil, nil, err
	}
	var roots []bundle.Root
	var records []app.AirgapBundleArtifact
	buildIDs := map[string]string{}
	for _, connection := range cfg.ComponentConfigConnections {
		buildID := published.ComponentBuildIDs[connection.ID]
		if buildID == "" {
			return logical, nil, nil, nil, fmt.Errorf("bundle has no pinned build for component %s", connection.ComponentName)
		}
		var build app.ComponentBuild
		if err := a.db.WithContext(ctx).Where(app.ComponentBuild{ID: buildID, OrgID: published.OrgID, ComponentConfigConnectionID: connection.ID}).First(&build).Error; err != nil {
			return logical, nil, nil, nil, fmt.Errorf("load pinned build %s for component %s: %w", buildID, connection.ComponentName, err)
		}
		desc, err := repo.Resolve(ctx, build.ID)
		if err != nil {
			return logical, nil, nil, nil, fmt.Errorf("resolve component build %s: %w", build.ID, err)
		}
		name := connection.ComponentName
		if name == "" {
			name = connection.ComponentID
		}
		configDigest := objectDigest(connection)
		artifact := bundleArtifact(desc)
		logical.Components = append(logical.Components, bundle.Component{Name: name, Type: string(connection.Type), ConfigDigest: configDigest, Source: bundle.Source{Digest: build.SourceDigest, RequestedRef: build.SourceRef}, Artifact: artifact})
		roots = append(roots, bundle.Root{Descriptor: desc, Source: repo})
		records = append(records, artifactRecord("component", name, artifact, currentApp.Repository.RepositoryURI, configDigest, connection.ID, ""))
		buildIDs["component:"+name] = build.ID
		if image, record := externalImageEntries(connection, name, artifact, currentApp.Repository.RepositoryURI, configDigest); image != nil {
			logical.Images = append(logical.Images, *image)
			records = append(records, *record)
		}
	}
	var sandboxBuild app.AppSandboxBuild
	if published.SandboxBuildID == "" {
		return logical, nil, nil, nil, fmt.Errorf("bundle has no pinned sandbox build")
	}
	if err := a.db.WithContext(ctx).Where(app.AppSandboxBuild{ID: published.SandboxBuildID, OrgID: published.OrgID, AppID: published.AppID, AppConfigID: cfg.ID, AppSandboxConfigID: cfg.SandboxConfig.ID}).First(&sandboxBuild).Error; err != nil {
		return logical, nil, nil, nil, fmt.Errorf("load pinned sandbox build %s: %w", published.SandboxBuildID, err)
	}
	sandboxDesc, err := repo.Resolve(ctx, sandboxBuild.ID)
	if err != nil {
		return logical, nil, nil, nil, fmt.Errorf("resolve sandbox build %s: %w", sandboxBuild.ID, err)
	}
	sandboxArtifact := bundleArtifact(sandboxDesc)
	sandboxConfigDigest := objectDigest(cfg.SandboxConfig)
	logical.Sandbox = &bundle.Sandbox{Type: cfg.SandboxConfig.Type, ConfigDigest: sandboxConfigDigest, Artifact: sandboxArtifact}
	roots = append(roots, bundle.Root{Descriptor: sandboxDesc, Source: repo})
	records = append(records, artifactRecord("sandbox", "sandbox", sandboxArtifact, currentApp.Repository.RepositoryURI, sandboxConfigDigest, "", cfg.SandboxConfig.ID))
	buildIDs["sandbox"] = sandboxBuild.ID

	for _, actionCfg := range cfg.ActionWorkflowConfigs {
		name := actionCfg.ActionWorkflow.Name
		if name == "" {
			name = actionCfg.ActionWorkflowID
		}
		// Envelope compilation excludes Git-sourced actions with a
		// qualification warning; skip them here too so the manifest
		// matches what the envelope actually ships.
		gitSourced := false
		for _, stepCfg := range actionCfg.Steps {
			if stepCfg.PublicGitVCSConfig != nil || stepCfg.ConnectedGithubVCSConfig != nil {
				gitSourced = true
				break
			}
		}
		if gitSourced {
			continue
		}
		action := bundle.Action{Name: name, ConfigDigest: objectDigest(actionCfg)}
		for _, stepCfg := range actionCfg.Steps {
			step := bundle.Step{Name: stepCfg.Name, Command: stepCfg.Command}
			if stepCfg.InlineContents != "" {
				step.InlineContentsDigest = digest.FromString(stepCfg.InlineContents).String()
				store, desc, err := packedArtifact(ctx, "application/vnd.nuon.airgap.action-source.v1", "text/x-shellscript", []byte(stepCfg.InlineContents))
				if err != nil {
					return logical, nil, nil, nil, err
				}
				artifact := bundleArtifact(desc)
				step.Artifact = &artifact
				roots = append(roots, bundle.Root{Descriptor: desc, Source: store})
				records = append(records, artifactRecord("action_step", name+"/"+stepCfg.Name, artifact, "inline", objectDigest(stepCfg), "", ""))
			}
			action.Steps = append(action.Steps, step)
		}
		logical.Actions = append(logical.Actions, action)
	}
	initScriptURL := cfg.RunnerConfig.InitScriptURL
	if initScriptURL == "" {
		initScriptURL = defaultAWSRunnerInitScriptURL
	}
	assets := []struct{ role, source, contentsHash, mediaType string }{
		{role: "runner", source: cfg.StackConfig.RunnerNestedTemplateURL, mediaType: stackAssetYAMLMediaType},
		{role: "vpc", source: cfg.StackConfig.VPCNestedTemplateURL, mediaType: stackAssetYAMLMediaType},
		{role: "init_script", source: initScriptURL, mediaType: stackAssetShellMediaType},
	}
	for _, custom := range cfg.StackConfig.CustomNestedStacks {
		source, err := customStackSource(a.cfg.AWSCloudFormationStackTemplateBaseURL, published.OrgID, published.AppID, custom.ContentsHash, custom.TemplateURL)
		if err != nil {
			return logical, nil, nil, nil, fmt.Errorf("resolve custom stack asset %s: %w", custom.Name, err)
		}
		assets = append(assets, struct{ role, source, contentsHash, mediaType string }{role: "custom:" + custom.Name, source: source, contentsHash: custom.ContentsHash, mediaType: stackAssetYAMLMediaType})
	}
	for _, asset := range assets {
		if asset.source == "" {
			continue
		}
		data, err := fetchStackAsset(ctx, asset.source, a.cfg.AWSCloudFormationStackTemplateBaseURL, maxFetchedAsset)
		if err != nil {
			return logical, nil, nil, nil, fmt.Errorf("fetch stack asset %s: %w", asset.role, err)
		}
		if asset.contentsHash != "" {
			actual := fmt.Sprintf("%x", sha256.Sum256(data))
			if !strings.EqualFold(actual, asset.contentsHash) {
				return logical, nil, nil, nil, fmt.Errorf("stack asset %s content hash mismatch: expected %s, got %s", asset.role, asset.contentsHash, actual)
			}
		}
		if asset.role == "runner" {
			if err := validateRunnerNestedTemplateAirgapCompatible(data); err != nil {
				return logical, nil, nil, nil, fmt.Errorf("stack asset %s: %w", asset.role, err)
			}
		}
		store, desc, err := packedArtifact(ctx, "application/vnd.nuon.airgap.stack.v1", asset.mediaType, data)
		if err != nil {
			return logical, nil, nil, nil, err
		}
		logical.StackAssets = append(logical.StackAssets, bundle.StackAsset{Role: asset.role, SourceURL: asset.source, Digest: desc.Digest.String(), MediaType: desc.MediaType, Size: desc.Size})
		roots = append(roots, bundle.Root{Descriptor: desc, Source: store})
		records = append(records, artifactRecord("stack_asset", asset.role, bundleArtifact(desc), asset.source, "", "", ""))
	}
	rootTemplate, rootTemplateSource, err := a.rootTemplateInputs(ctx, published.OrgID, referenceInstallID, airgap.VirtualInstallID(cfg.ID), cfg)
	if err != nil {
		return logical, nil, nil, nil, err
	}
	rootStore, rootDesc, err := packedArtifact(ctx, "application/vnd.nuon.airgap.stack.v1", stackAssetJSONMediaType, rootTemplate)
	if err != nil {
		return logical, nil, nil, nil, err
	}
	logical.StackAssets = append(logical.StackAssets, bundle.StackAsset{Role: rootTemplateAssetRole, SourceURL: rootTemplateSource, Digest: rootDesc.Digest.String(), MediaType: rootDesc.MediaType, Size: rootDesc.Size})
	roots = append(roots, bundle.Root{Descriptor: rootDesc, Source: rootStore})
	records = append(records, artifactRecord("stack_asset", rootTemplateAssetRole, bundleArtifact(rootDesc), rootTemplateSource, "", "", ""))
	if a.cfg.AirgapRunnerBinaryURL != "" {
		runnerRoots, runnerRecords, runner, err := a.runnerInputs(ctx, logical.Target)
		if err != nil {
			return logical, nil, nil, nil, err
		}
		logical.Runner = runner
		roots = append(roots, runnerRoots...)
		records = append(records, runnerRecords...)
	}
	return logical, uniqueRoots(roots), records, map[string]any{"app_config_id": cfg.ID, "build_ids": buildIDs}, nil
}

func (a *Activities) runnerInputs(ctx context.Context, target bundle.Target) ([]bundle.Root, []app.AirgapBundleArtifact, *bundle.Runner, error) {
	data, err := fetchRunnerBinary(ctx, a.cfg.AirgapRunnerBinaryURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetch runner binary: %w", err)
	}
	store, desc, err := packedArtifact(ctx, bundle.RunnerBinaryArtifactType, bundle.RunnerBinaryMediaType, data)
	if err != nil {
		return nil, nil, nil, err
	}
	binaryArtifact := bundleArtifact(desc)
	runner := &bundle.Runner{Version: a.cfg.RunnerContainerImageTag, SourceURL: a.cfg.AirgapRunnerBinaryURL, Binary: &binaryArtifact}
	roots := []bundle.Root{{Descriptor: desc, Source: store}}
	records := []app.AirgapBundleArtifact{artifactRecord("runner_binary", "runner", binaryArtifact, a.cfg.AirgapRunnerBinaryURL, "", "", "")}

	imageRepoCfg := &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypePublicOCI, Repository: a.cfg.RunnerContainerImageURL}
	imageRepo, err := oci.GetRepo(ctx, imageRepoCfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open runner image repo: %w", err)
	}
	imageDesc, err := resolvePlatformImage(ctx, imageRepo, a.cfg.RunnerContainerImageTag, target)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("resolve runner image %s:%s: %w", a.cfg.RunnerContainerImageURL, a.cfg.RunnerContainerImageTag, err)
	}
	imageArtifact := bundleArtifact(imageDesc)
	if imageDesc.Platform != nil {
		imageArtifact.PlatformOS = imageDesc.Platform.OS
		imageArtifact.PlatformArchitecture = imageDesc.Platform.Architecture
	}
	runner.Image = &bundle.Image{Name: "runner", Repository: a.cfg.RunnerContainerImageURL + ":" + a.cfg.RunnerContainerImageTag, Artifact: imageArtifact}
	roots = append(roots, bundle.Root{Descriptor: imageDesc, Source: imageRepo})
	records = append(records, artifactRecord("runner_image", "runner", imageArtifact, a.cfg.RunnerContainerImageURL, "", "", ""))
	return roots, records, runner, nil
}

func resolvePlatformImage(ctx context.Context, repo registry.Repository, tag string, target bundle.Target) (ocispec.Descriptor, error) {
	desc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	if desc.MediaType != ocispec.MediaTypeImageIndex && desc.MediaType != "application/vnd.docker.distribution.manifest.list.v2+json" {
		return desc, nil
	}
	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	defer rc.Close()
	body, err := io.ReadAll(io.LimitReader(rc, maxFetchedAsset))
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	var index ocispec.Index
	if err := json.Unmarshal(body, &index); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("decode image index: %w", err)
	}
	for _, m := range index.Manifests {
		if m.Platform != nil && m.Platform.OS == target.OS && m.Platform.Architecture == target.Architecture {
			return m, nil
		}
	}
	return ocispec.Descriptor{}, fmt.Errorf("image index has no %s/%s manifest", target.OS, target.Architecture)
}

const maxRunnerBinary = int64(1 << 30)

func fetchRunnerBinary(ctx context.Context, source string) ([]byte, error) {
	u, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "file":
		f, err := os.Open(u.Path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return readAtMost(f, maxRunnerBinary)
	case "https", "http":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
		if err != nil {
			return nil, err
		}
		resp, err := assetHTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d fetching runner binary", resp.StatusCode)
		}
		return readAtMost(resp.Body, maxRunnerBinary)
	default:
		return nil, fmt.Errorf("unsupported runner binary URL scheme %q", u.Scheme)
	}
}

func readAtMost(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("runner binary exceeds %d byte limit", limit)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("runner binary is empty")
	}
	return data, nil
}

func customStackSource(baseURL, orgID, appID, contentsHash, templateURL string) (string, error) {
	if len(contentsHash) != sha256.Size*2 {
		return "", fmt.Errorf("custom stack content hash must be a 64-character SHA-256 digest")
	}
	for _, c := range contentsHash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return "", fmt.Errorf("custom stack content hash must be a 64-character SHA-256 digest")
		}
	}
	return cloudformation.CustomNestedStackTemplateURL(baseURL, orgID, appID, contentsHash, templateURL), nil
}

func uniqueRoots(roots []bundle.Root) []bundle.Root {
	unique := make([]bundle.Root, 0, len(roots))
	seen := make(map[digest.Digest]struct{}, len(roots))
	for _, root := range roots {
		if _, ok := seen[root.Descriptor.Digest]; ok {
			continue
		}
		seen[root.Descriptor.Digest] = struct{}{}
		unique = append(unique, root)
	}
	return unique
}

func objectDigest(v any) string { b, _ := json.Marshal(v); return digest.FromBytes(b).String() }
func bundleArtifact(d ocispec.Descriptor) bundle.Artifact {
	return bundle.Artifact{MediaType: d.MediaType, Digest: d.Digest.String(), Size: d.Size}
}
func artifactRecord(kind, name string, a bundle.Artifact, repository, configDigest, connectionID, sandboxID string) app.AirgapBundleArtifact {
	return app.AirgapBundleArtifact{Kind: kind, LogicalName: name, ComponentConfigConnectionID: connectionID, AppSandboxConfigID: sandboxID, ConfigDigest: configDigest, Repository: repository, Digest: a.Digest, MediaType: a.MediaType, Size: a.Size}
}

func externalImageEntries(connection app.ComponentConfigConnection, name string, artifact bundle.Artifact, repository, configDigest string) (*bundle.Image, *app.AirgapBundleArtifact) {
	if connection.Type != app.ComponentTypeExternalImage || connection.ExternalImageComponentConfig == nil {
		return nil, nil
	}
	image := bundle.Image{Name: name, Repository: connection.ExternalImageComponentConfig.ImageURL, Artifact: artifact}
	record := artifactRecord("image", name, artifact, repository, configDigest, connection.ID, "")
	return &image, &record
}

func packedArtifact(ctx context.Context, artifactType, layerType string, data []byte) (*memory.Store, ocispec.Descriptor, error) {
	store := memory.New()
	layer, err := oras.PushBytes(ctx, store, layerType, data)
	if err != nil {
		return nil, ocispec.Descriptor{}, err
	}
	desc, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, artifactType, oras.PackManifestOptions{Layers: []ocispec.Descriptor{layer}, ManifestAnnotations: map[string]string{ocispec.AnnotationCreated: time.Unix(0, 0).UTC().Format(time.RFC3339)}})
	return store, desc, err
}

func fetchLimited(ctx context.Context, source string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, err
	}
	resp, err := assetHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("content exceeds %d bytes", limit)
	}
	return b, nil
}

func fetchStackAsset(ctx context.Context, source, configuredBaseURL string, limit int64) ([]byte, error) {
	u, err := validateStackAssetURL(source, configuredBaseURL)
	if err != nil {
		return nil, err
	}
	return fetchLimited(ctx, u.String(), limit)
}

func validateStackAssetURL(source, configuredBaseURL string) (*url.URL, error) {
	u, err := url.Parse(source)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil {
		return nil, fmt.Errorf("stack asset URL must be an absolute HTTPS URL")
	}
	allowedHost := false
	if base, baseErr := url.Parse(configuredBaseURL); baseErr == nil && base.Scheme == "https" && strings.EqualFold(base.Hostname(), u.Hostname()) {
		allowedHost = true
	}
	host := strings.ToLower(u.Hostname())
	if host == "s3.amazonaws.com" || (strings.HasSuffix(host, ".amazonaws.com") && (strings.HasPrefix(host, "s3.") || strings.Contains(host, ".s3."))) {
		allowedHost = true
	}
	if host == "raw.githubusercontent.com" {
		allowedHost = true
	}
	if !allowedHost {
		return nil, fmt.Errorf("stack asset host %q is not an approved immutable asset host", u.Hostname())
	}
	return u, nil
}
