package bundleupgrade

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2run"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2state"
	"github.com/nuonco/nuon/pkg/runner/oci/bundle"
)

const StagedCandidateKey = "day2/staged-candidate.json"

type ObjectStore interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type DeploymentContext struct {
	Bucket           string
	StatePrefix      string
	BucketPrefix     string
	Region           string
	Image            string
	ECRRegistry      string
	DeploymentID     string
	BundleKey        string
	BundleURI        string
	StackOutputsKey  string
	InstallInputsURI string
	InitTemplate     string
}

type Input struct {
	ArchivePath string
	ArchiveName string
	Deployment  DeploymentContext
	Store       ObjectStore
	Now         func() time.Time
	Progress    func(Progress)
}

type Progress struct {
	Phase  string `json:"phase"`
	Detail string `json:"detail"`
}

type Result struct {
	Candidate       day2.BundleCandidate
	CandidateKey    string
	ArchiveKey      string
	RootTemplateKey string
}

func Stage(ctx context.Context, in Input) (*Result, error) {
	if in.ArchivePath == "" || in.Store == nil {
		return nil, fmt.Errorf("archive path and object store are required")
	}
	d := in.Deployment
	if d.Bucket == "" || d.StatePrefix == "" || d.Region == "" || d.BundleKey == "" || d.BundleURI == "" || d.Image == "" {
		return nil, fmt.Errorf("bucket, state prefix, region, bundle key, bundle URI, and image are required")
	}
	d.StatePrefix = strings.Trim(d.StatePrefix, "/") + "/"
	runnerStatePrefix := d.StatePrefix + day2state.RunnerNamespace
	controlStatePrefix := d.StatePrefix + day2state.ControlNamespace
	reportProgress(in.Progress, "extracting", "Extracting the uploaded bundle")
	workdir, err := os.MkdirTemp("", "nuon-bundle-upgrade-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(workdir)
	archiveInfo, err := os.Stat(in.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("stat candidate archive: %w", err)
	}
	archive, err := os.Open(in.ArchivePath)
	if err != nil {
		return nil, fmt.Errorf("open candidate archive: %w", err)
	}
	checksum, err := bundle.Extract(workdir, archive)
	archive.Close()
	if err != nil {
		return nil, err
	}
	b, err := bundle.Open(ctx, workdir)
	if err != nil {
		return nil, err
	}
	reportProgress(in.Progress, "verifying", "Verifying bundle metadata and content digests")
	if err := bundle.VerifyBlobs(workdir); err != nil {
		return nil, err
	}
	if b.Manifest.SchemaVersion != 2 {
		return nil, fmt.Errorf("upgrade candidate must be a v2 bundle (got schema version %d)", b.Manifest.SchemaVersion)
	}
	if len(b.PlanEnvelope) == 0 {
		return nil, fmt.Errorf("upgrade candidate has no plan envelope")
	}
	envelope, err := airgap.Parse(b.PlanEnvelope)
	if err != nil {
		return nil, fmt.Errorf("parse candidate plan envelope: %w", err)
	}
	effectiveInstallID := envelope.InstallID
	deploymentInstallID := ""
	if d.DeploymentID != "" {
		effectiveInstallID, err = airgap.DeploymentInstallID(envelope.InstallID, d.DeploymentID)
		if err != nil {
			return nil, err
		}
		deploymentInstallID = effectiveInstallID
	}
	reportProgress(in.Progress, "loading-active", "Reading the active installation state")
	active, catalogDeploymentID, catalogDigest, statusInstallID, err := readIdentity(ctx, in.Store, d.Bucket, runnerStatePrefix)
	if err != nil {
		active, catalogDeploymentID, catalogDigest, statusInstallID, err = readIdentity(ctx, in.Store, d.Bucket, d.StatePrefix)
	}
	if err != nil {
		return nil, err
	}
	if needsCanonicalHydration(active) {
		reportProgress(in.Progress, "loading-active", "Downloading and verifying the active bundle for comparison")
		active, err = hydrateActiveBundle(ctx, in.Store, d.Bucket, d.BundleKey, active)
		if err != nil {
			return nil, err
		}
	}
	if err := validateIdentity(active, catalogDeploymentID, catalogDigest, statusInstallID, effectiveInstallID); err != nil {
		return nil, err
	}
	now := time.Now
	if in.Now != nil {
		now = in.Now
	}
	reportProgress(in.Progress, "comparing", "Comparing the uploaded bundle with the active bundle")
	candidateInfo := day2run.BundleInfoFromManifest(effectiveInstallID, day2.EnvelopeDigest(b.PlanEnvelope), b.Manifest, time.Time{})
	candidateInfo.SchemaVersion = b.Manifest.SchemaVersion
	candidateInfo.ArchiveDigest = "sha256:" + checksum
	changes := day2.CompareBundleContents(active, *candidateInfo)
	for _, change := range changes {
		if change.Kind == day2.BundleContentKindComponent && change.Change == day2.BundleChangeRemoved {
			return nil, fmt.Errorf("candidate removes component %q; component removal is not supported", change.Name)
		}
		if change.Kind == day2.BundleContentKindSandbox && change.Change != day2.BundleChangeUnchanged && (change.Change != day2.BundleChangeChanged || change.Name != "terraform") {
			return nil, fmt.Errorf("candidate sandbox transition %q for %q is not supported", change.Change, change.Name)
		}
	}
	if active.BundleDigest == candidateInfo.BundleDigest {
		return nil, fmt.Errorf("candidate bundle %s is already active", candidateInfo.BundleDigest)
	}
	prefix := normalizePrefix(d.BucketPrefix)
	if d.DeploymentID != "" {
		prefix += d.DeploymentID + "/"
	}
	candidatePrefix := prefix + "stack/candidates/" + safeDigest(candidateInfo.BundleDigest) + "/"
	keys := makeKeys(candidatePrefix)
	imageURL, imageTag, registry, err := splitImageRef(d.Image)
	if err != nil {
		return nil, err
	}
	if d.ECRRegistry != "" {
		registry = d.ECRRegistry
	}
	runnerID := "airgap-" + effectiveInstallID
	initTemplate := d.InitTemplate
	if initTemplate == "" {
		for _, asset := range b.Manifest.StackAssets {
			if asset.Role != "init_script" {
				continue
			}
			layer, err := contentLayer(workdir, asset)
			if err != nil {
				return nil, err
			}
			raw, err := os.ReadFile(blobPath(workdir, layer.Digest.String()))
			if err != nil {
				return nil, err
			}
			initTemplate = string(raw)
			break
		}
	}
	if initTemplate == "" {
		return nil, fmt.Errorf("upgrade candidate has no init script template")
	}
	initScript, err := renderInit(initTemplate, initData{Bucket: d.Bucket, Prefix: candidatePrefix, Region: d.Region, ImageURL: imageURL, ImageTag: imageTag, ECRRegistry: registry, BundleURI: d.BundleURI, StatePrefix: strings.TrimSuffix(d.StatePrefix, "/"), RunnerID: runnerID, StackOutputsURI: "s3://" + d.Bucket + "/" + d.StackOutputsKey, InstallInputsURI: d.InstallInputsURI, DeploymentID: d.DeploymentID})
	if err != nil {
		return nil, err
	}
	reportProgress(in.Progress, "publishing-assets", "Publishing runner bootstrap assets")
	if err := putBytes(ctx, in.Store, d.Bucket, keys.Init, "text/x-shellscript", initScript); err != nil {
		return nil, err
	}
	var runner bytes.Buffer
	if err := b.ExtractRunnerBinary(ctx, &runner); err != nil {
		return nil, err
	}
	if err := putBytes(ctx, in.Store, d.Bucket, keys.Runner, "application/octet-stream", runner.Bytes()); err != nil {
		return nil, err
	}
	reused := map[string]bool{}
	for _, change := range changes {
		if change.Kind == day2.BundleContentKindStackAsset && change.Change == day2.BundleChangeUnchanged {
			reused[change.Name] = true
		}
		if change.Kind == day2.BundleContentKindRunnerBinary && change.Change != day2.BundleChangeUnchanged {
			delete(reused, "runner")
			delete(reused, "init_script")
		}
	}
	for _, asset := range b.Manifest.StackAssets {
		reportProgress(in.Progress, "publishing-assets", "Publishing stack asset "+asset.Role)
		layer, err := contentLayer(workdir, asset)
		if err != nil {
			return nil, err
		}
		raw, err := os.ReadFile(blobPath(workdir, layer.Digest.String()))
		if err != nil {
			return nil, err
		}
		if asset.Role == "runner" {
			raw, err = rewriteRunner(raw)
			if err != nil {
				return nil, err
			}
		}
		raw = substituteInstallID(raw, envelope.InstallID, deploymentInstallID)
		if err := putBytes(ctx, in.Store, d.Bucket, candidatePrefix+"stack/"+asset.Role, layer.MediaType, raw); err != nil {
			return nil, err
		}
	}
	root, err := rewriteRoot(workdir, b.Manifest, keys, d.Bucket, d.Region, candidatePrefix, prefix, registry, runnerID, envelope.InstallID, deploymentInstallID, reused)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, fmt.Errorf("upgrade candidate has no install stack template")
	}
	if err := putBytes(ctx, in.Store, d.Bucket, keys.RootTemplate, "application/json", root); err != nil {
		return nil, err
	}
	archiveKey := prefix + "bundle/candidates/" + safeDigest(candidateInfo.BundleDigest) + ".tar.zst"
	stagedAt := now().UTC()
	archiveName := in.ArchiveName
	if archiveName == "" {
		archiveName = filepath.Base(in.ArchivePath)
	}
	pointer := day2.BundleCandidate{SchemaVersion: 1, PreviousDigest: active.BundleDigest, StagedAt: stagedAt, ArchiveName: filepath.Base(archiveName), ArchiveSize: archiveInfo.Size(), Bundle: *candidateInfo, Changes: changes, Deployment: &day2.BundleDeploymentAssets{StackTemplateURL: customerAssetURL(d.Bucket, d.Region, keys.RootTemplate), CandidateBundleKey: archiveKey, TargetBundleKey: d.BundleKey}}
	raw, err := json.MarshalIndent(pointer, "", "  ")
	if err != nil {
		return nil, err
	}
	recordKey := day2.CandidateStageKey(candidateInfo.BundleDigest, stagedAt)
	if err := stageCandidate(ctx, in.Store, d.Bucket, controlStatePrefix, in.ArchivePath, archiveKey, recordKey, raw, in.Progress); err != nil {
		return nil, err
	}
	return &Result{Candidate: pointer, CandidateKey: recordKey, ArchiveKey: archiveKey, RootTemplateKey: keys.RootTemplate}, nil
}

func needsCanonicalHydration(info day2.BundleInfo) bool {
	for _, content := range info.Contents {
		if content.Kind == day2.BundleContentKindAction && content.ActionDefinition == nil {
			return true
		}
		if content.Kind == day2.BundleContentKindComponent && content.ComponentDefinition == nil {
			return true
		}
	}
	return false
}

func hydrateActiveBundle(ctx context.Context, store ObjectStore, bucket, key string, recorded day2.BundleInfo) (day2.BundleInfo, error) {
	out, err := store.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
	if err != nil {
		return day2.BundleInfo{}, fmt.Errorf("read active bundle s3://%s/%s: %w", bucket, key, err)
	}
	defer out.Body.Close()
	dir, err := os.MkdirTemp("", "nuon-active-bundle-*")
	if err != nil {
		return day2.BundleInfo{}, err
	}
	defer os.RemoveAll(dir)
	archiveDigest, err := bundle.Extract(dir, out.Body)
	if err != nil {
		return day2.BundleInfo{}, fmt.Errorf("extract active bundle: %w", err)
	}
	opened, err := bundle.Open(ctx, dir)
	if err != nil {
		return day2.BundleInfo{}, fmt.Errorf("open active bundle: %w", err)
	}
	if err := bundle.VerifyBlobs(dir); err != nil {
		return day2.BundleInfo{}, fmt.Errorf("verify active bundle: %w", err)
	}
	hydrated := day2run.BundleInfoFromManifest(recorded.DeploymentID, day2.EnvelopeDigest(opened.PlanEnvelope), opened.Manifest, recorded.ActivatedAt)
	hydrated.SchemaVersion = opened.Manifest.SchemaVersion
	hydrated.ArchiveDigest = "sha256:" + archiveDigest
	return *hydrated, nil
}

func stageCandidate(ctx context.Context, store ObjectStore, bucket, statePrefix, archivePath, archiveKey, recordKey string, raw []byte, progress func(Progress)) error {
	reportProgress(progress, "publishing-archive", "Publishing the candidate bundle archive")
	if err := putFile(ctx, store, bucket, archiveKey, "application/zstd", archivePath); err != nil {
		return err
	}
	reportProgress(progress, "recording", "Recording the staged bundle and its diff")
	if err := putBytes(ctx, store, bucket, statePrefix+recordKey, "application/json", raw); err != nil {
		return err
	}
	return putBytes(ctx, store, bucket, statePrefix+StagedCandidateKey, "application/json", raw)
}

func reportProgress(report func(Progress), phase, detail string) {
	if report != nil {
		report(Progress{Phase: phase, Detail: detail})
	}
}

func get(ctx context.Context, store ObjectStore, bucket, key string, dst any) error {
	out, err := store.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
	if err != nil {
		return fmt.Errorf("read s3://%s/%s: %w", bucket, key, err)
	}
	defer out.Body.Close()
	if err := json.NewDecoder(out.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode s3://%s/%s: %w", bucket, key, err)
	}
	return nil
}

func readIdentity(ctx context.Context, store ObjectStore, bucket, prefix string) (day2.BundleInfo, string, string, string, error) {
	var active day2.BundleInfo
	if err := get(ctx, store, bucket, prefix+"day2/bundle.json", &active); err != nil {
		return active, "", "", "", err
	}
	var catalog struct {
		DeploymentID string `json:"deployment_id"`
		BundleDigest string `json:"bundle_digest"`
	}
	if err := get(ctx, store, bucket, prefix+day2.CatalogKey, &catalog); err != nil {
		return active, "", "", "", err
	}
	var status struct {
		InstallID string `json:"install_id"`
	}
	if err := get(ctx, store, bucket, prefix+"status.json", &status); err != nil {
		return active, "", "", "", err
	}
	return active, catalog.DeploymentID, catalog.BundleDigest, status.InstallID, nil
}

func validateIdentity(active day2.BundleInfo, catalogDeploymentID, catalogDigest, statusInstallID, installID string) error {
	if active.DeploymentID == "" || active.BundleDigest == "" || statusInstallID == "" || catalogDigest == "" {
		return fmt.Errorf("active bundle, status, or catalog is missing deployment identity")
	}
	for _, check := range []struct{ field, active, candidate string }{{"deployment ID", active.DeploymentID, installID}, {"catalog deployment ID", catalogDeploymentID, installID}, {"status install ID", statusInstallID, installID}, {"catalog bundle digest", catalogDigest, active.BundleDigest}} {
		if check.active != "" && check.candidate != check.active {
			return fmt.Errorf("candidate %s %q does not match active deployment %q", check.field, check.candidate, check.active)
		}
	}
	return nil
}

type keys struct{ Runner, Init, RootTemplate string }

func makeKeys(prefix string) keys {
	return keys{prefix + "bootstrap/runner", prefix + "bootstrap/init.sh", prefix + "stack/root-template.json"}
}
func normalizePrefix(s string) string {
	s = strings.Trim(s, "/")
	if s == "" {
		return ""
	}
	return s + "/"
}
func safeDigest(s string) string { return strings.ReplaceAll(s, ":", "-") }
func customerAssetURL(bucket, region, key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)
}
func blobPath(dir, digest string) string {
	return filepath.Join(dir, "blobs", "sha256", strings.TrimPrefix(digest, "sha256:"))
}

func contentLayer(dir string, asset bundle.StackAsset) (ocispec.Descriptor, error) {
	raw, err := os.ReadFile(blobPath(dir, asset.Digest))
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return ocispec.Descriptor{}, err
	}
	if len(manifest.Layers) != 1 {
		return ocispec.Descriptor{}, fmt.Errorf("stack asset %s manifest has %d layers, expected 1", asset.Role, len(manifest.Layers))
	}
	return manifest.Layers[0], nil
}

func putBytes(ctx context.Context, store ObjectStore, bucket, key, mediaType string, raw []byte) error {
	_, err := store.PutObject(ctx, &s3.PutObjectInput{Bucket: &bucket, Key: &key, Body: bytes.NewReader(raw), ContentType: aws.String(mediaType)})
	if err != nil {
		return fmt.Errorf("write s3://%s/%s: %w", bucket, key, err)
	}
	return nil
}
func putFile(ctx context.Context, store ObjectStore, bucket, key, mediaType, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = store.PutObject(ctx, &s3.PutObjectInput{Bucket: &bucket, Key: &key, Body: f, ContentType: aws.String(mediaType)})
	if err != nil {
		return fmt.Errorf("write s3://%s/%s: %w", bucket, key, err)
	}
	return nil
}

type initData struct{ Bucket, Prefix, Region, ImageURL, ImageTag, ECRRegistry, BundleURI, StatePrefix, StackOutputsURI, InstallInputsURI, RunnerID, DeploymentID string }

func renderInit(source string, data initData) ([]byte, error) {
	t, err := template.New("init").Option("missingkey=error").Parse(source)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
func splitImageRef(ref string) (string, string, string, error) {
	slash, colon := strings.IndexByte(ref, '/'), strings.LastIndexByte(ref, ':')
	if slash <= 0 || colon <= slash+1 || colon == len(ref)-1 {
		return "", "", "", fmt.Errorf("image must be a full tagged ECR image reference")
	}
	registry := ref[:slash]
	if !strings.Contains(registry, ".dkr.ecr.") {
		return "", "", "", fmt.Errorf("image must use an ECR registry")
	}
	return ref[:colon], ref[colon+1:], registry, nil
}

const connectedInitFetch = "curl ${RunnerInitScriptUrl} | bash"
const airgapInitFetch = "for i in $(seq 1 60); do aws s3 cp ${RunnerInitScriptUrl} /tmp/nuon-init.sh && break; sleep 10; done; bash /tmp/nuon-init.sh"

func rewriteRunner(raw []byte) ([]byte, error) {
	if !bytes.Contains(raw, []byte(connectedInitFetch)) {
		return nil, fmt.Errorf("runner stack template does not contain expected init fetch command")
	}
	return bytes.ReplaceAll(raw, []byte(connectedInitFetch), []byte(airgapInitFetch)), nil
}
func substituteInstallID(raw []byte, frozen, deployed string) []byte {
	if frozen == "" || deployed == "" || frozen == deployed {
		return raw
	}
	return bytes.ReplaceAll(raw, []byte(frozen), []byte(deployed))
}

func rewriteRoot(dir string, manifest bundle.LogicalManifest, keys keys, bucket, region, assetPrefix, accessPrefix, registry, runnerID, frozenID, deployedID string, reused map[string]bool) ([]byte, error) {
	var root *bundle.StackAsset
	for i := range manifest.StackAssets {
		if manifest.StackAssets[i].Role == "root" {
			root = &manifest.StackAssets[i]
			break
		}
	}
	if root == nil {
		return nil, nil
	}
	layer, err := contentLayer(dir, *root)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(blobPath(dir, layer.Digest.String()))
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	replacements := map[string]string{}
	for _, asset := range manifest.StackAssets {
		if asset.Role == "root" || asset.SourceURL == "" {
			continue
		}
		p := assetPrefix
		if reused[asset.Role] {
			p = accessPrefix
		}
		target := customerAssetURL(bucket, region, p+"stack/"+asset.Role)
		if asset.Role == "init_script" {
			key := keys.Init
			if reused[asset.Role] {
				key = accessPrefix + "bootstrap/init.sh"
			}
			target = "s3://" + bucket + "/" + key
		}
		replacements[asset.SourceURL] = target
		if base, _, ok := strings.Cut(asset.SourceURL, "#"); ok {
			replacements[base] = target
		}
	}
	rewriteValues(doc, replacements)
	params, _ := doc["Parameters"].(map[string]any)
	for name, value := range params {
		if strings.HasPrefix(name, "EnableBreakglass") {
			if spec, ok := value.(map[string]any); ok {
				spec["Default"] = "true"
			}
		}
	}
	resources, ok := doc["Resources"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("root stack template has no Resources section")
	}
	if _, ok := resources["RunnerAutoScalingGroup"]; !ok {
		return nil, fmt.Errorf("root stack template has no RunnerAutoScalingGroup resource")
	}
	account, _, _ := strings.Cut(registry, ".")
	parts := strings.Split(registry, ".")
	ecrRegion := ""
	if len(parts) > 3 {
		ecrRegion = parts[3]
	}
	if account == "" || ecrRegion == "" {
		return nil, fmt.Errorf("unable to derive account and region from ECR registry %q", registry)
	}
	resources["NuonAirgapAssetAccess"] = map[string]any{"Type": "AWS::IAM::Policy", "Properties": map[string]any{"PolicyName": "nuon-airgap-asset-access", "Roles": []any{map[string]any{"Fn::GetAtt": []any{"RunnerAutoScalingGroup", "Outputs.RunnerInstanceRole"}}}, "PolicyDocument": map[string]any{"Version": "2012-10-17", "Statement": []any{map[string]any{"Effect": "Allow", "Action": []any{"s3:GetObject"}, "Resource": fmt.Sprintf("arn:aws:s3:::%s/%s*", bucket, accessPrefix)}, map[string]any{"Effect": "Allow", "Action": []any{"s3:ListBucket"}, "Resource": "arn:aws:s3:::" + bucket}, map[string]any{"Effect": "Allow", "Action": []any{"s3:PutObject"}, "Resource": fmt.Sprintf("arn:aws:s3:::%s/%sstate/*", bucket, accessPrefix)}, map[string]any{"Effect": "Allow", "Action": []any{"ecr:GetAuthorizationToken"}, "Resource": "*"}, map[string]any{"Effect": "Allow", "Action": []any{"ecr:BatchCheckLayerAvailability", "ecr:BatchGetImage", "ecr:GetDownloadUrlForLayer", "ecr:DescribeRepositories", "ecr:DescribeImages", "ecr:ListImages", "ecr:InitiateLayerUpload", "ecr:UploadLayerPart", "ecr:CompleteLayerUpload", "ecr:PutImage", "ecr:CreateRepository", "ecr:TagResource"}, "Resource": fmt.Sprintf("arn:aws:ecr:%s:%s:repository/*", ecrRegion, account)}, map[string]any{"Effect": "Allow", "Action": []any{"logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents", "logs:DescribeLogStreams"}, "Resource": fmt.Sprintf("arn:aws:logs:*:*:log-group:runner-%s*", runnerID)}}}}}
	out, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	out = substituteInstallID(out, frozenID, deployedID)
	out = bytes.ReplaceAll(out, []byte("__NUON_AIRGAP_STACK_region__"), []byte(region))
	return out, nil
}
func rewriteValues(value any, replacements map[string]string) any {
	switch node := value.(type) {
	case string:
		if v, ok := replacements[node]; ok {
			return v
		}
		return node
	case map[string]any:
		for k, v := range node {
			node[k] = rewriteValues(v, replacements)
		}
	case []any:
		for i, v := range node {
			node[i] = rewriteValues(v, replacements)
		}
	}
	return value
}
