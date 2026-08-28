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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gorm.io/gorm"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	customermanaged "github.com/nuonco/nuon/pkg/customer_managed"
	bundle "github.com/nuonco/nuon/pkg/customer_managed/bundle"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/runner/oci"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	customermanagedapp "github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
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
	PackageID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 180m
func (a *Activities) PublishBundle(ctx context.Context, req *PublishBundleRequest) error {
	var published app.ReleasePackage
	if err := a.db.WithContext(ctx).Preload("Release").Where(app.ReleasePackage{ID: req.PackageID}).First(&published).Error; err != nil {
		return fmt.Errorf("load release package: %w", err)
	}
	release := &published.Release
	if published.ManifestDigest != "" && published.OCIRootDigest != "" && published.ArchiveChecksum != "" && published.OCIIndexDigest != "" {
		var replicas []app.ReleasePackageReplica
		if err := a.db.WithContext(ctx).Where(app.ReleasePackageReplica{PackageID: published.ID, OrgID: published.OrgID}).Find(&replicas).Error; err != nil {
			return fmt.Errorf("load release package replicas: %w", err)
		}
		for _, replica := range replicas {
			if replica.VerifiedAt != nil {
				return a.db.WithContext(ctx).Model(&app.ReleasePackage{ID: published.ID}).Updates(app.ReleasePackage{Status: app.ReleasePackageStatusActive, StatusDescription: "package published and verified"}).Error
			}
		}
	}
	cfg, err := a.appsHelpers.GetCustomerManagedAppConfig(ctx, release.OrgID, release.AppID, release.AppConfigID)
	if err != nil {
		return err
	}
	report := customermanagedapp.Qualify(cfg, published.TargetPlatform)
	if !report.Qualified {
		return fmt.Errorf("app config does not qualify for portable bundle export")
	}
	envelope, err := customermanagedapp.CompilePlanEnvelope(ctx, a.db, a.v, release.OrgID, release.AppID, cfg, release.SandboxBuildID, release.ComponentBuildIDs, release.Runbooks, &report)
	if err != nil {
		return fmt.Errorf("compile plan envelope: %w", err)
	}
	logical, roots, artifacts, provenance, err := a.bundleInputs(ctx, release, published.TargetPlatform, cfg, "")
	if err != nil {
		return err
	}
	logical.Release = bundle.ReleaseIdentity{ID: release.ID, Digest: release.SemanticDigest}
	logical.Package = bundle.PackageIdentity{ID: published.ID, Digest: published.PackageDigest, Format: published.Format, Target: published.TargetPlatform}
	pins, err := a.bundlePins(ctx, release, cfg)
	if err != nil {
		return err
	}
	if err := customermanagedapp.RewriteEnvelopeForBundle(ctx, a.db, envelope, pins); err != nil {
		return fmt.Errorf("rewrite plan envelope to pinned bundle members: %w", err)
	}
	logical.Runbooks, err = canonicalRunbookDefinitions(ctx, a.db, cfg, envelope)
	if err != nil {
		return fmt.Errorf("canonicalize runbooks: %w", err)
	}
	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal plan envelope: %w", err)
	}
	reportJSON, _ := json.Marshal(report)
	provenanceJSON, _ := json.Marshal(provenance)
	var sourceArchiveJSON json.RawMessage
	if release.DefinitionsBlob != nil && release.DefinitionsBlob.IsSet() {
		raw, err := release.DefinitionsBlob.Get(blobstore.WithBlobService(ctx, a.blobSvc))
		if err != nil {
			return fmt.Errorf("load release source archive: %w", err)
		}
		if !json.Valid([]byte(raw)) {
			return fmt.Errorf("release source archive is not valid JSON")
		}
		sourceArchiveJSON = json.RawMessage(raw)
		sourceDigest := digest.FromBytes(sourceArchiveJSON).String()
		artifacts = append(artifacts, app.ReleasePackageMember{
			Kind: "source_archive", LogicalName: "source", Digest: sourceDigest,
			MediaType: bundle.SourceArchiveMediaType, Size: int64(len(sourceArchiveJSON)),
		})
	}
	f, err := os.CreateTemp("", "nuon-customer-managed-*.tar.zst")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	generateOpts := bundle.GenerateOptions{MaxContentBytes: maxBundleContent, MaxBlobBytes: maxBundleBlob}
	if a.store.Configured() {
		generateOpts.BlobSink = func(dgst digest.Digest, data []byte) error {
			return a.store.PublishBlob(ctx, release.OrgID, dgst.Encoded(), data)
		}
	}
	result, err := bundle.GenerateWithOptions(ctx, f, logical, bundle.Documents{
		Provenance: provenanceJSON, QualificationReport: reportJSON, PlanEnvelope: envelopeJSON, SourceArchive: sourceArchiveJSON,
	}, roots, generateOpts)
	if err != nil {
		return fmt.Errorf("generate bundle: %w", err)
	}
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	indexDigest := digest.FromBytes(result.Index)
	if a.store.Configured() {
		if err := a.store.PublishBlob(ctx, release.OrgID, indexDigest.Encoded(), result.Index); err != nil {
			return fmt.Errorf("publish bundle index blob: %w", err)
		}
	}
	replica, err := a.store.Publish(ctx, transport.PublishRequest{Body: f, Size: stat.Size(), SHA256: result.TransportSHA256})
	if err != nil {
		return fmt.Errorf("publish bundle: %w", err)
	}
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := app.ReleasePackage{ManifestDigest: result.ManifestDescriptor.Digest.String(), PlanDigest: digest.FromBytes(envelopeJSON).String(), OCIRootDigest: result.BundleDescriptor.Digest.String(), OCIIndexDigest: indexDigest.String(), ArchiveChecksum: result.TransportSHA256, ArchiveSize: stat.Size(), Status: app.ReleasePackageStatusActive, StatusDescription: "package published and verified"}
		if err := tx.Model(&app.ReleasePackage{ID: published.ID}).Updates(updates).Error; err != nil {
			return err
		}
		for i := range artifacts {
			artifacts[i].PackageID = published.ID
			artifacts[i].OrgID = release.OrgID
		}
		if err := tx.Where(app.ReleasePackageMember{PackageID: published.ID}).Delete(&app.ReleasePackageMember{}).Error; err != nil {
			return err
		}
		if len(artifacts) > 0 {
			if err := tx.Create(&artifacts).Error; err != nil {
				return err
			}
		}
		verified := replica.VerifiedAt
		r := app.ReleasePackageReplica{PackageID: published.ID, OrgID: release.OrgID, Provider: replica.Provider, Region: replica.Region, StorageRef: replica.StorageRef, StorageVersion: replica.StorageVersion, ArchiveChecksum: replica.TransportChecksum, Size: replica.Size, VerifiedAt: &verified}
		return tx.Create(&r).Error
	})
	if err != nil {
		return err
	}
	return nil
}

func (a *Activities) bundlePins(ctx context.Context, release *app.AppRelease, cfg *app.AppConfig) (customermanagedapp.BundlePins, error) {
	pins := customermanagedapp.BundlePins{SandboxBuildID: release.SandboxBuildID, ComponentBuildIDs: map[string]string{}}
	for _, connection := range cfg.ComponentConfigConnections {
		buildID := release.ComponentBuildIDs[connection.ID]
		if buildID == "" {
			return pins, fmt.Errorf("bundle has no pinned build for component %s", connection.ComponentName)
		}
		pins.ComponentBuildIDs[connection.ComponentID] = buildID
	}
	var currentApp app.App
	if err := a.db.WithContext(ctx).Preload("Repository").Where(app.App{ID: release.AppID, OrgID: release.OrgID}).First(&currentApp).Error; err != nil {
		return pins, err
	}
	// No auth block on purpose: the customer-managed runner serves this source from the
	// bundle by tag, so the registry is a fail-closed placeholder.
	pins.SandboxRegistry = &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypeECR, Repository: currentApp.Repository.RepositoryURI, Region: currentApp.Repository.Region}
	return pins, nil
}

func (a *Activities) bundleInputs(ctx context.Context, release *app.AppRelease, targetPlatform string, cfg *app.AppConfig, referenceInstallID string) (bundle.LogicalManifest, []bundle.Root, []app.ReleasePackageMember, map[string]any, error) {
	target, err := bundleTarget(targetPlatform)
	if err != nil {
		return bundle.LogicalManifest{}, nil, nil, nil, err
	}
	logical := bundle.LogicalManifest{SchemaVersion: bundle.CurrentSchemaVersion, Target: target}
	var currentApp app.App
	if err := a.db.WithContext(ctx).Preload("Repository").Where(app.App{ID: release.AppID, OrgID: release.OrgID}).First(&currentApp).Error; err != nil {
		return logical, nil, nil, nil, err
	}
	repoCfg := &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypeECR, Repository: currentApp.Repository.RepositoryURI, Region: currentApp.Repository.Region, ECRAuth: &credentials.Config{Region: currentApp.Repository.Region, AssumeRole: &credentials.AssumeRoleConfig{RoleARN: a.cfg.ManagementIAMRoleARN, SessionName: "customer_managed_bundle"}}}
	repo, err := oci.GetRepo(ctx, repoCfg)
	if err != nil {
		return logical, nil, nil, nil, err
	}
	var roots []bundle.Root
	var records []app.ReleasePackageMember
	buildIDs := map[string]string{}
	for _, connection := range cfg.ComponentConfigConnections {
		buildID := release.ComponentBuildIDs[connection.ID]
		if buildID == "" {
			return logical, nil, nil, nil, fmt.Errorf("bundle has no pinned build for component %s", connection.ComponentName)
		}
		var build app.ComponentBuild
		if err := a.db.WithContext(ctx).Where(app.ComponentBuild{ID: buildID, OrgID: release.OrgID, ComponentConfigConnectionID: connection.ID}).First(&build).Error; err != nil {
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
		definition, err := canonicalComponentDefinition(connection, cfg.ComponentConfigConnections)
		if err != nil {
			return logical, nil, nil, nil, fmt.Errorf("canonicalize component %s: %w", name, err)
		}
		configDigest := objectDigest(definition)
		artifact := bundleArtifact(desc)
		totalSize, err := bundle.TotalSize(ctx, repo, desc)
		if err != nil {
			return logical, nil, nil, nil, fmt.Errorf("compute total size for component %s: %w", name, err)
		}
		logical.Components = append(logical.Components, bundle.Component{Name: name, Type: string(connection.Type), ConfigDigest: configDigest, Definition: definition, Source: bundle.Source{Digest: build.SourceDigest, RequestedRef: build.SourceRef}, Artifact: artifact})
		roots = append(roots, bundle.Root{Descriptor: desc, Source: repo})
		componentRecord := artifactRecord("component", name, artifact, currentApp.Repository.RepositoryURI, configDigest, connection.ID, "")
		componentRecord.ComponentID = connection.ComponentID
		componentRecord.Size = totalSize
		records = append(records, componentRecord)
		buildIDs["component:"+name] = build.ID
		if image, record := externalImageEntries(connection, name, artifact, currentApp.Repository.RepositoryURI, configDigest); image != nil {
			logical.Images = append(logical.Images, *image)
			record.Size = totalSize
			records = append(records, *record)
		}
	}
	if release.SandboxBuildID == "" {
		return logical, nil, nil, nil, fmt.Errorf("bundle has no pinned sandbox build")
	}
	sandboxBuild, err := customermanagedapp.LoadPinnedSandboxBuild(ctx, a.db, release.OrgID, release.AppID, release.SandboxBuildID)
	if err != nil {
		return logical, nil, nil, nil, fmt.Errorf("load pinned sandbox build %s: %w", release.SandboxBuildID, err)
	}
	sandboxDesc, err := repo.Resolve(ctx, sandboxBuild.ID)
	if err != nil {
		return logical, nil, nil, nil, fmt.Errorf("resolve sandbox build %s: %w", sandboxBuild.ID, err)
	}
	sandboxArtifact := bundleArtifact(sandboxDesc)
	sandboxConfigDigest := objectDigest(cfg.SandboxConfig)
	sandboxTotalSize, err := bundle.TotalSize(ctx, repo, sandboxDesc)
	if err != nil {
		return logical, nil, nil, nil, fmt.Errorf("compute total size for sandbox: %w", err)
	}
	logical.Sandbox = &bundle.Sandbox{Type: cfg.SandboxConfig.Type, ConfigDigest: sandboxConfigDigest, Artifact: sandboxArtifact}
	roots = append(roots, bundle.Root{Descriptor: sandboxDesc, Source: repo})
	sandboxRecord := artifactRecord("sandbox", "sandbox", sandboxArtifact, currentApp.Repository.RepositoryURI, sandboxConfigDigest, "", cfg.SandboxConfig.ID)
	sandboxRecord.Size = sandboxTotalSize
	records = append(records, sandboxRecord)
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
		definition := canonicalActionDefinition(actionCfg, cfg.ComponentConfigConnections)
		action := bundle.Action{Name: name, ConfigDigest: objectDigest(definition), Definition: &definition}
		for _, stepCfg := range actionCfg.Steps {
			step := bundle.Step{Name: stepCfg.Name, Command: stepCfg.Command}
			if stepCfg.InlineContents != "" {
				step.InlineContentsDigest = digest.FromString(stepCfg.InlineContents).String()
				store, desc, err := packedArtifact(ctx, "application/vnd.nuon.customer_managed.action-source.v1", "text/x-shellscript", []byte(stepCfg.InlineContents))
				if err != nil {
					return logical, nil, nil, nil, err
				}
				artifact := bundleArtifact(desc)
				step.Artifact = &artifact
				roots = append(roots, bundle.Root{Descriptor: desc, Source: store})
				stepTotalSize, err := bundle.TotalSize(ctx, store, desc)
				if err != nil {
					return logical, nil, nil, nil, fmt.Errorf("compute total size for action step %s/%s: %w", name, stepCfg.Name, err)
				}
				stepRecord := artifactRecord("action_step", name+"/"+stepCfg.Name, artifact, "inline", objectDigest(stepCfg), "", "")
				stepRecord.ActionWorkflowID = actionCfg.ActionWorkflowID
				stepRecord.Size = stepTotalSize
				records = append(records, stepRecord)
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
		source, err := customStackSource(a.cfg.AWSCloudFormationStackTemplateBaseURL, release.OrgID, release.AppID, custom.ContentsHash, custom.TemplateURL)
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
			if err := validateRunnerNestedTemplateCustomerManagedCompatible(data); err != nil {
				return logical, nil, nil, nil, fmt.Errorf("stack asset %s: %w", asset.role, err)
			}
		}
		store, desc, err := packedArtifact(ctx, "application/vnd.nuon.customer_managed.stack.v1", asset.mediaType, data)
		if err != nil {
			return logical, nil, nil, nil, err
		}
		logical.StackAssets = append(logical.StackAssets, bundle.StackAsset{Role: asset.role, SourceURL: asset.source, Digest: desc.Digest.String(), MediaType: desc.MediaType, Size: desc.Size})
		roots = append(roots, bundle.Root{Descriptor: desc, Source: store})
		assetTotalSize, err := bundle.TotalSize(ctx, store, desc)
		if err != nil {
			return logical, nil, nil, nil, fmt.Errorf("compute total size for stack asset %s: %w", asset.role, err)
		}
		assetRecord := artifactRecord("stack_asset", asset.role, bundleArtifact(desc), asset.source, "", "", "")
		assetRecord.Size = assetTotalSize
		records = append(records, assetRecord)
	}
	rootTemplate, rootTemplateSource, err := a.rootTemplateInputs(ctx, release.OrgID, referenceInstallID, customermanagedapp.VirtualInstallID(release.AppID), cfg, release.Runtime.RunnerImageTag)
	if err != nil {
		return logical, nil, nil, nil, err
	}
	rootStore, rootDesc, err := packedArtifact(ctx, "application/vnd.nuon.customer_managed.stack.v1", stackAssetJSONMediaType, rootTemplate)
	if err != nil {
		return logical, nil, nil, nil, err
	}
	logical.StackAssets = append(logical.StackAssets, bundle.StackAsset{Role: rootTemplateAssetRole, SourceURL: rootTemplateSource, Digest: rootDesc.Digest.String(), MediaType: rootDesc.MediaType, Size: rootDesc.Size})
	roots = append(roots, bundle.Root{Descriptor: rootDesc, Source: rootStore})
	rootTemplateTotalSize, err := bundle.TotalSize(ctx, rootStore, rootDesc)
	if err != nil {
		return logical, nil, nil, nil, fmt.Errorf("compute total size for stack asset %s: %w", rootTemplateAssetRole, err)
	}
	rootTemplateRecord := artifactRecord("stack_asset", rootTemplateAssetRole, bundleArtifact(rootDesc), rootTemplateSource, "", "", "")
	rootTemplateRecord.Size = rootTemplateTotalSize
	records = append(records, rootTemplateRecord)
	platformRuntime, ok := release.Runtime.Platforms[targetPlatform]
	if !ok {
		return logical, nil, nil, nil, fmt.Errorf("release has no runtime artifacts for target platform %s", targetPlatform)
	}
	if platformRuntime.PortalBinaryURL != "" {
		portalRoot, portalRecord, portalAsset, err := a.portalInputs(ctx, platformRuntime.PortalBinaryURL)
		if err != nil {
			return logical, nil, nil, nil, err
		}
		logical.StackAssets = append(logical.StackAssets, portalAsset)
		roots = append(roots, portalRoot)
		records = append(records, portalRecord)
	}
	if platformRuntime.RunnerBinaryURL != "" {
		runnerRoots, runnerRecords, runner, err := a.runnerInputs(ctx, logical.Target, release.Runtime, platformRuntime.RunnerBinaryURL)
		if err != nil {
			return logical, nil, nil, nil, err
		}
		logical.Runner = runner
		roots = append(roots, runnerRoots...)
		records = append(records, runnerRecords...)
	}
	return logical, uniqueRoots(roots), records, map[string]any{"app_config_id": cfg.ID, "build_ids": buildIDs}, nil
}

func (a *Activities) portalInputs(ctx context.Context, sourceURL string) (bundle.Root, app.ReleasePackageMember, bundle.StackAsset, error) {
	data, err := fetchBinary(ctx, sourceURL)
	if err != nil {
		return bundle.Root{}, app.ReleasePackageMember{}, bundle.StackAsset{}, fmt.Errorf("fetch portal binary: %w", err)
	}
	store, desc, err := packedArtifact(ctx, "application/vnd.nuon.customer_managed.portal-binary.v1", bundle.RunnerBinaryMediaType, data)
	if err != nil {
		return bundle.Root{}, app.ReleasePackageMember{}, bundle.StackAsset{}, err
	}
	artifact := bundleArtifact(desc)
	totalSize, err := bundle.TotalSize(ctx, store, desc)
	if err != nil {
		return bundle.Root{}, app.ReleasePackageMember{}, bundle.StackAsset{}, fmt.Errorf("compute total size for portal binary: %w", err)
	}
	record := artifactRecord("portal_binary", "portal", artifact, sourceURL, "", "", "")
	record.Size = totalSize
	asset := bundle.StackAsset{Role: "portal_binary", SourceURL: sourceURL, Digest: desc.Digest.String(), MediaType: desc.MediaType, Size: desc.Size}
	return bundle.Root{Descriptor: desc, Source: store}, record, asset, nil
}

func (a *Activities) runnerInputs(ctx context.Context, target bundle.Target, runtime app.AppReleaseRuntime, runnerBinaryURL string) ([]bundle.Root, []app.ReleasePackageMember, *bundle.Runner, error) {
	data, err := fetchBinary(ctx, runnerBinaryURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetch runner binary: %w", err)
	}
	store, desc, err := packedArtifact(ctx, bundle.RunnerBinaryArtifactType, bundle.RunnerBinaryMediaType, data)
	if err != nil {
		return nil, nil, nil, err
	}
	binaryArtifact := bundleArtifact(desc)
	runner := &bundle.Runner{Version: runtime.RunnerImageTag, SourceURL: runnerBinaryURL, Binary: &binaryArtifact}
	roots := []bundle.Root{{Descriptor: desc, Source: store}}
	binaryTotalSize, err := bundle.TotalSize(ctx, store, desc)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("compute total size for runner binary: %w", err)
	}
	binaryRecord := artifactRecord("runner_binary", "runner", binaryArtifact, runnerBinaryURL, "", "", "")
	binaryRecord.Size = binaryTotalSize
	records := []app.ReleasePackageMember{binaryRecord}

	imageRepoCfg := &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypePublicOCI, Repository: runtime.RunnerImageURL}
	imageRepo, err := oci.GetRepo(ctx, imageRepoCfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open runner image repo: %w", err)
	}
	imageDesc, err := resolvePlatformImage(ctx, imageRepo, runtime.RunnerImageTag, target)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("resolve runner image %s:%s: %w", runtime.RunnerImageURL, runtime.RunnerImageTag, err)
	}
	if err := validateRunnerImageBundleSchema(ctx, imageRepo, imageDesc, bundle.CurrentSchemaVersion); err != nil {
		return nil, nil, nil, fmt.Errorf("qualify runner image %s:%s: %w", runtime.RunnerImageURL, runtime.RunnerImageTag, err)
	}
	imageArtifact := bundleArtifact(imageDesc)
	if imageDesc.Platform != nil {
		imageArtifact.PlatformOS = imageDesc.Platform.OS
		imageArtifact.PlatformArchitecture = imageDesc.Platform.Architecture
	}
	runner.Image = &bundle.Image{Name: "runner", Repository: runtime.RunnerImageURL + ":" + runtime.RunnerImageTag, Artifact: imageArtifact}
	roots = append(roots, bundle.Root{Descriptor: imageDesc, Source: imageRepo})
	imageTotalSize, err := bundle.TotalSize(ctx, imageRepo, imageDesc)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("compute total size for runner image: %w", err)
	}
	imageRecord := artifactRecord("runner_image", "runner", imageArtifact, runtime.RunnerImageURL, "", "", "")
	imageRecord.Size = imageTotalSize
	records = append(records, imageRecord)
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

func validateRunnerImageBundleSchema(ctx context.Context, repo registry.Repository, imageDesc ocispec.Descriptor, schemaVersion int) error {
	rc, err := repo.Fetch(ctx, imageDesc)
	if err != nil {
		return fmt.Errorf("fetch manifest: %w", err)
	}
	body, err := io.ReadAll(io.LimitReader(rc, maxFetchedAsset))
	_ = rc.Close()
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}
	rc, err = repo.Fetch(ctx, manifest.Config)
	if err != nil {
		return fmt.Errorf("fetch config: %w", err)
	}
	body, err = io.ReadAll(io.LimitReader(rc, maxFetchedAsset))
	_ = rc.Close()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	var image ocispec.Image
	if err := json.Unmarshal(body, &image); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}
	return validateRunnerImageBundleSchemaLabels(image.Config.Labels, schemaVersion)
}

func validateRunnerImageBundleSchemaLabels(labels map[string]string, schemaVersion int) error {
	const minLabel = "io.nuon.customer_managed.bundle-schema.min"
	const maxLabel = "io.nuon.customer_managed.bundle-schema.max"
	minVersion, minErr := strconv.Atoi(labels[minLabel])
	maxVersion, maxErr := strconv.Atoi(labels[maxLabel])
	if minErr != nil || maxErr != nil || minVersion < 1 || maxVersion < minVersion {
		return fmt.Errorf("missing or invalid %s/%s compatibility labels", minLabel, maxLabel)
	}
	if schemaVersion < minVersion || schemaVersion > maxVersion {
		return fmt.Errorf("bundle schema %d is outside runner-supported range %d-%d", schemaVersion, minVersion, maxVersion)
	}
	return nil
}

const maxBinaryAsset = int64(1 << 30)

func fetchBinary(ctx context.Context, source string) ([]byte, error) {
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
		return readAtMost(f, maxBinaryAsset)
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
			return nil, fmt.Errorf("unexpected status %d fetching binary", resp.StatusCode)
		}
		return readAtMost(resp.Body, maxBinaryAsset)
	default:
		return nil, fmt.Errorf("unsupported binary URL scheme %q", u.Scheme)
	}
}

func readAtMost(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("binary exceeds %d byte limit", limit)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("binary is empty")
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

func canonicalComponentDefinition(connection app.ComponentConfigConnection, connections []app.ComponentConfigConnection) (bundle.ComponentDefinition, error) {
	return customermanagedapp.CanonicalComponentDefinition(connection, connections)
}

func canonicalActionDefinition(actionCfg app.ActionWorkflowConfig, connections []app.ComponentConfigConnection) bundle.ActionDefinition {
	return customermanagedapp.CanonicalActionDefinition(actionCfg, connections)
}

func objectDigest(value any) string { return customermanagedapp.ObjectDigest(value) }

func canonicalRunbookDefinitions(ctx context.Context, db *gorm.DB, cfg *app.AppConfig, envelope *customermanaged.Envelope) ([]bundle.Runbook, error) {
	result := canonicalEnvelopeRunbookDefinitions(envelope)
	byName := make(map[string]bundle.Runbook, len(result))
	for _, runbook := range result {
		byName[runbook.Name] = runbook
	}
	var configs []app.RunbookConfig
	if err := db.WithContext(ctx).
		Preload("Runbook").Preload("Steps").Preload("Inputs").
		Where(app.RunbookConfig{OrgID: cfg.OrgID, AppConfigID: cfg.ID}).Find(&configs).Error; err != nil {
		return nil, err
	}
	for _, runbookConfig := range configs {
		name := runbookConfig.Runbook.Name
		if name == "" {
			name = runbookConfig.RunbookID
		}
		definition := customermanagedapp.CanonicalRunbookDefinition(runbookConfig, cfg.ActionWorkflowConfigs)
		byName[name] = bundle.Runbook{Name: name, ConfigDigest: objectDigest(definition), Definition: definition}
	}
	result = result[:0]
	for _, runbook := range byName {
		result = append(result, runbook)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func canonicalEnvelopeRunbookDefinitions(envelope *customermanaged.Envelope) []bundle.Runbook {
	references := make(map[string]string, len(envelope.Actions)+len(envelope.Drift))
	for _, action := range envelope.Actions {
		references[action.ID] = "action:" + action.Name
	}
	for _, drift := range envelope.Drift {
		references[drift.ID] = "drift:" + drift.ComponentName
	}
	result := make([]bundle.Runbook, 0, len(envelope.Runbooks))
	for _, runbook := range envelope.Runbooks {
		definition := bundle.RunbookDefinition{Steps: make([]bundle.RunbookStepDefinition, 0, len(runbook.Steps))}
		for _, step := range runbook.Steps {
			reference := references[step.RefID]
			if reference == "" {
				reference = step.RefID
			}
			definition.Steps = append(definition.Steps, bundle.RunbookStepDefinition{
				Kind: step.Kind, Reference: reference, Component: step.Component,
			})
		}
		result = append(result, bundle.Runbook{Name: runbook.Name, ConfigDigest: objectDigest(definition), Definition: definition})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}
func bundleArtifact(d ocispec.Descriptor) bundle.Artifact {
	return bundle.Artifact{MediaType: d.MediaType, Digest: d.Digest.String(), Size: d.Size}
}

func bundleTarget(targetPlatform string) (bundle.Target, error) {
	parts := strings.Split(targetPlatform, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return bundle.Target{}, fmt.Errorf("invalid target platform %q: expected os/architecture", targetPlatform)
	}
	return bundle.Target{OS: parts[0], Architecture: parts[1]}, nil
}

func artifactRecord(kind, name string, a bundle.Artifact, repository, configDigest, connectionID, sandboxID string) app.ReleasePackageMember {
	return app.ReleasePackageMember{Kind: kind, LogicalName: name, ComponentConfigConnectionID: connectionID, AppSandboxConfigID: sandboxID, ConfigDigest: configDigest, Repository: repository, Digest: a.Digest, MediaType: a.MediaType, Size: a.Size}
}

func externalImageEntries(connection app.ComponentConfigConnection, name string, artifact bundle.Artifact, repository, configDigest string) (*bundle.Image, *app.ReleasePackageMember) {
	if connection.Type != app.ComponentTypeExternalImage || connection.ExternalImageComponentConfig == nil {
		return nil, nil
	}
	image := bundle.Image{Name: name, Repository: connection.ExternalImageComponentConfig.ImageURL, Artifact: artifact}
	record := artifactRecord("image", name, artifact, repository, configDigest, connection.ID, "")
	record.ComponentID = connection.ComponentID
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
