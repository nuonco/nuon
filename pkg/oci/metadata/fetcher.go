package metadata

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/nuonco/nuon/pkg/oci/dockerhub"
)

const (
	SBOMMediaTypeSPDX      = "application/spdx+json"
	SBOMMediaTypeCycloneDX = "application/vnd.cyclonedx+json"
	SignatureMediaType     = "application/vnd.dev.cosign.simplesigning.v1+json"

	ArtifactTypeSBOM      = "application/vnd.oci.artifact.sbom.v1+json"
	ArtifactTypeSignature = "application/vnd.dev.cosign.artifact.sig.v1+json"

	MediaTypeImageIndex     = "application/vnd.oci.image.index.v1+json"
	MediaTypeDockerManifest = "application/vnd.docker.distribution.manifest.list.v2+json"

	AnnotationReferenceType   = "vnd.docker.reference.type"
	AnnotationReferenceDigest = "vnd.docker.reference.digest"
	AnnotationPredicateType   = "in-toto.io/predicate-type"
	ReferenceTypeAttestation  = "attestation-manifest"

	MediaTypeInToto = "application/vnd.in-toto+json"

	MediaTypeCosignSignature   = "application/vnd.dev.cosign.simplesigning.v1+json"
	MediaTypeDSSEEnvelope      = "application/vnd.dsse.envelope.v1+json"
	MediaTypeOCIImageManifest  = "application/vnd.oci.image.manifest.v1+json"
	MediaTypeDockerImageConfig = "application/vnd.oci.image.config.v1+json"
)

var ErrNotIndex = errors.New("not an image index")

type RegistryAuth struct {
	ServerAddress string
	Username      string
	Password      string
}

type FetchGuardrails struct {
	MaxBlobBytes         int64
	MaxTotalBytes        int64
	MaxAttestations      int
	MaxLayersPerManifest int
}

func DefaultGuardrails() FetchGuardrails {
	return FetchGuardrails{
		MaxBlobBytes:         10 * 1024 * 1024,
		MaxTotalBytes:        10 * 1024 * 1024,
		MaxAttestations:      10,
		MaxLayersPerManifest: 5,
	}
}

type FetchOptions struct {
	Image  string
	Tag    string
	Auth   *RegistryAuth
	Digest string

	IncludeIndex                bool
	IncludeAttestationManifests bool
	IncludeAttestationLayers    bool

	Platform string

	Guardrails *FetchGuardrails
}

func newAuthClient(serverAddr string, regAuth *RegistryAuth) *auth.Client {
	client := &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
	}
	if regAuth != nil && regAuth.Username != "" {
		client.Credential = auth.StaticCredential(serverAddr, auth.Credential{
			Username: regAuth.Username,
			Password: regAuth.Password,
		})
	}
	return client
}

func FetchImageMetadata(ctx context.Context, opts *FetchOptions) (*ImageMetadata, error) {
	normalizedImage := dockerhub.NormalizeReference(opts.Image)
	repo, err := remote.NewRepository(normalizedImage)
	if err != nil {
		return nil, fmt.Errorf("unable to create repository client: %w", err)
	}

	serverAddr := ""
	if opts.Auth != nil {
		serverAddr = strings.TrimPrefix(opts.Auth.ServerAddress, "https://")
		serverAddr = strings.TrimPrefix(serverAddr, "http://")
	}
	if serverAddr == "" {
		parts := strings.SplitN(opts.Image, "/", 2)
		if len(parts) > 0 {
			serverAddr = parts[0]
		}
	}
	repo.Client = newAuthClient(serverAddr, opts.Auth)

	tag := opts.Tag
	if tag == "" {
		tag = "latest"
	}

	guardrails := opts.Guardrails
	if guardrails == nil {
		g := DefaultGuardrails()
		guardrails = &g
	}

	desc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve image tag %s: %w", tag, err)
	}

	result := &ImageMetadata{
		Image:  opts.Image,
		Tag:    tag,
		Digest: desc.Digest.String(),
		Signed: false,
		SBOM:   nil,
	}

	if opts.IncludeIndex || opts.IncludeAttestationManifests {
		index, err := fetchImageIndex(ctx, repo, desc)
		if err != nil {
			if !errors.Is(err, errdef.ErrNotFound) && !errors.Is(err, ErrNotIndex) {
				return nil, fmt.Errorf("unable to fetch image index: %w", err)
			}
		} else {
			result.Index = index

			if opts.IncludeAttestationManifests && index != nil {
				attestationManifests, err := fetchAttestationManifests(ctx, repo, index, opts, guardrails)
				if err != nil {
					return nil, fmt.Errorf("unable to fetch attestation manifests: %w", err)
				}
				result.AttestationManifests = attestationManifests
			}
		}
	}

	referrers, err := fetchReferrers(ctx, repo, desc)
	if err != nil {
		if !errors.Is(err, errdef.ErrNotFound) {
			return nil, fmt.Errorf("unable to fetch referrers: %w", err)
		}
	}

	for _, ref := range referrers {
		switch {
		case isSBOMArtifact(ref):
			format := detectSBOMFormat(ref.ArtifactType, ref.MediaType)
			result.SBOM = &SBOM{
				Present: true,
				Format:  format,
			}
		case isSignatureArtifact(ref):
			result.Signed = true
			result.Signatures = append(result.Signatures, Signature{
				Algorithm: ref.MediaType,
			})
		case isAttestationArtifact(ref):
			result.Attestations = append(result.Attestations, Attestation{
				Type: ref.ArtifactType,
			})
		}
	}

	// why: Fallback: Try Cosign tag-based discovery if no signatures/attestations found via referrers
	// This handles registries like GHCR that don't support the OCI 1.1 Referrers API
	if !result.Signed && len(result.Attestations) == 0 {
		cosignResult, err := fetchCosignTagBasedArtifacts(ctx, repo, desc, guardrails)
		if err == nil && cosignResult != nil {
			if cosignResult.Signed {
				result.Signed = true
				result.Signatures = append(result.Signatures, cosignResult.Signatures...)
			}
			if len(cosignResult.Attestations) > 0 {
				result.Attestations = append(result.Attestations, cosignResult.Attestations...)
			}
			if cosignResult.SBOM != nil && result.SBOM == nil {
				result.SBOM = cosignResult.SBOM
			}
		}
	}

	if result.SBOM == nil {
		if sbom := detectSBOMFromAttestationManifests(result.AttestationManifests); sbom != nil {
			result.SBOM = sbom
		}
	}

	return result, nil
}

func fetchImageIndex(ctx context.Context, repo *remote.Repository, desc v1.Descriptor) (*ImageIndex, error) {
	if !isIndexMediaType(desc.MediaType) {
		return nil, fmt.Errorf("%w: media type is %s", ErrNotIndex, desc.MediaType)
	}

	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch index: %w", err)
	}
	defer rc.Close()

	rawJSON, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("unable to read index content: %w", err)
	}

	var ociIndex v1.Index
	if err := json.Unmarshal(rawJSON, &ociIndex); err != nil {
		return nil, fmt.Errorf("unable to parse index: %w", err)
	}

	index := &ImageIndex{
		Digest:    desc.Digest.String(),
		MediaType: desc.MediaType,
		RawJSON:   rawJSON,
		Manifests: make([]ManifestEntry, 0, len(ociIndex.Manifests)),
	}

	for _, m := range ociIndex.Manifests {
		entry := ManifestEntry{
			Digest:      m.Digest.String(),
			MediaType:   m.MediaType,
			Size:        m.Size,
			Annotations: m.Annotations,
		}

		if m.Platform != nil {
			entry.Platform = &Platform{
				OS:           m.Platform.OS,
				Architecture: m.Platform.Architecture,
				Variant:      m.Platform.Variant,
			}
		}

		if refType, ok := m.Annotations[AnnotationReferenceType]; ok && refType == ReferenceTypeAttestation {
			entry.IsAttestation = true
		}

		index.Manifests = append(index.Manifests, entry)
	}

	return index, nil
}

func isIndexMediaType(mediaType string) bool {
	return mediaType == MediaTypeImageIndex || mediaType == MediaTypeDockerManifest
}

func fetchAttestationManifests(
	ctx context.Context,
	repo *remote.Repository,
	index *ImageIndex,
	opts *FetchOptions,
	guardrails *FetchGuardrails,
) ([]AttestationManifest, error) {
	var manifests []AttestationManifest
	var totalBytes int64
	attestationCount := 0

	platformFilter := parsePlatformFilter(opts.Platform)

	for _, entry := range index.Manifests {
		if !entry.IsAttestation {
			continue
		}

		if attestationCount >= guardrails.MaxAttestations {
			break
		}

		if platformFilter != nil && entry.Platform != nil {
			if !matchesPlatform(entry.Platform, platformFilter) {
				continue
			}
		}

		manifest, bytesRead, err := fetchAttestationManifest(ctx, repo, entry, opts, guardrails, totalBytes)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch attestation manifest %s: %w", entry.Digest, err)
		}

		totalBytes += bytesRead
		if totalBytes > guardrails.MaxTotalBytes {
			break
		}

		manifests = append(manifests, *manifest)
		attestationCount++
	}

	return manifests, nil
}

func parsePlatformFilter(platform string) *Platform {
	if platform == "" {
		return nil
	}

	parts := strings.Split(platform, "/")
	if len(parts) < 2 {
		return nil
	}

	p := &Platform{
		OS:           parts[0],
		Architecture: parts[1],
	}
	if len(parts) > 2 {
		p.Variant = parts[2]
	}
	return p
}

func matchesPlatform(actual, filter *Platform) bool {
	if filter.OS != "" && actual.OS != filter.OS {
		return false
	}
	if filter.Architecture != "" && actual.Architecture != filter.Architecture {
		return false
	}
	if filter.Variant != "" && actual.Variant != filter.Variant {
		return false
	}
	return true
}

func fetchAttestationManifest(
	ctx context.Context,
	repo *remote.Repository,
	entry ManifestEntry,
	opts *FetchOptions,
	guardrails *FetchGuardrails,
	currentTotalBytes int64,
) (*AttestationManifest, int64, error) {
	desc := v1.Descriptor{
		Digest:    digestFromString(entry.Digest),
		MediaType: entry.MediaType,
		Size:      entry.Size,
	}

	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to fetch manifest: %w", err)
	}
	defer rc.Close()

	rawJSON, err := io.ReadAll(rc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to read manifest content: %w", err)
	}

	bytesRead := int64(len(rawJSON))

	var ociManifest v1.Manifest
	if err := json.Unmarshal(rawJSON, &ociManifest); err != nil {
		return nil, bytesRead, fmt.Errorf("unable to parse manifest: %w", err)
	}

	manifest := &AttestationManifest{
		Digest:      entry.Digest,
		MediaType:   entry.MediaType,
		Platform:    entry.Platform,
		Annotations: entry.Annotations,
		RawJSON:     rawJSON,
	}

	if refDigest, ok := entry.Annotations[AnnotationReferenceDigest]; ok {
		manifest.RefDigest = refDigest
	}

	if opts.IncludeAttestationLayers {
		layers, layerBytes, err := fetchAttestationLayers(ctx, repo, ociManifest.Layers, guardrails, currentTotalBytes+bytesRead)
		if err != nil {
			return nil, bytesRead, fmt.Errorf("unable to fetch attestation layers: %w", err)
		}
		manifest.Layers = layers
		bytesRead += layerBytes
	} else {
		for _, l := range ociManifest.Layers {
			if len(manifest.Layers) >= guardrails.MaxLayersPerManifest {
				break
			}
			layer := AttestationLayer{
				Digest:    l.Digest.String(),
				MediaType: l.MediaType,
				Size:      l.Size,
			}
			if predicateType, ok := l.Annotations[AnnotationPredicateType]; ok {
				layer.PredicateType = predicateType
			}
			manifest.Layers = append(manifest.Layers, layer)
		}
	}

	return manifest, bytesRead, nil
}

func fetchAttestationLayers(
	ctx context.Context,
	repo *remote.Repository,
	layers []v1.Descriptor,
	guardrails *FetchGuardrails,
	currentTotalBytes int64,
) ([]AttestationLayer, int64, error) {
	if len(layers) == 0 {
		return nil, 0, nil
	}

	numLayers := len(layers)
	if numLayers > guardrails.MaxLayersPerManifest {
		numLayers = guardrails.MaxLayersPerManifest
	}

	results := make([]AttestationLayer, numLayers)
	bytesReadPerLayer := make([]int64, numLayers)

	g, gCtx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	var estimatedBytes int64

	for i := 0; i < numLayers; i++ {
		i := i
		l := layers[i]

		g.Go(func() error {
			mu.Lock()
			wouldExceedLimit := currentTotalBytes+estimatedBytes+l.Size > guardrails.MaxTotalBytes
			if !wouldExceedLimit {
				estimatedBytes += l.Size
			}
			mu.Unlock()

			if wouldExceedLimit {
				layer := AttestationLayer{
					Digest:    l.Digest.String(),
					MediaType: l.MediaType,
					Size:      l.Size,
					Truncated: true,
				}
				if predicateType, ok := l.Annotations[AnnotationPredicateType]; ok {
					layer.PredicateType = predicateType
				}
				results[i] = layer
				return nil
			}

			layer, layerBytes, err := fetchAttestationLayer(gCtx, repo, l, guardrails)
			if err != nil {
				mu.Lock()
				estimatedBytes -= l.Size
				mu.Unlock()
				return fmt.Errorf("unable to fetch layer %s: %w", l.Digest.String(), err)
			}

			bytesReadPerLayer[i] = layerBytes
			results[i] = *layer
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	var totalBytesRead int64
	for _, b := range bytesReadPerLayer {
		totalBytesRead += b
	}

	return results, totalBytesRead, nil
}

func fetchAttestationLayer(
	ctx context.Context,
	repo *remote.Repository,
	desc v1.Descriptor,
	guardrails *FetchGuardrails,
) (*AttestationLayer, int64, error) {
	layer := &AttestationLayer{
		Digest:    desc.Digest.String(),
		MediaType: desc.MediaType,
		Size:      desc.Size,
	}

	if predicateType, ok := desc.Annotations[AnnotationPredicateType]; ok {
		layer.PredicateType = predicateType
	}

	if desc.Size > guardrails.MaxBlobBytes {
		layer.Truncated = true
		return layer, 0, nil
	}

	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to fetch layer: %w", err)
	}
	defer rc.Close()

	rawJSON, err := io.ReadAll(io.LimitReader(rc, guardrails.MaxBlobBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("unable to read layer content: %w", err)
	}

	bytesRead := int64(len(rawJSON))
	layer.RawJSON = rawJSON

	decoded, err := decodeInTotoStatement(rawJSON)
	if err == nil && decoded != nil {
		layer.Decoded = decoded
		if layer.PredicateType == "" && decoded.PredicateType != "" {
			layer.PredicateType = decoded.PredicateType
		}
	}

	return layer, bytesRead, nil
}

func decodeInTotoStatement(data []byte) (*InTotoStatement, error) {
	var statement InTotoStatement
	if err := json.Unmarshal(data, &statement); err == nil && statement.Type != "" && statement.PredicateType != "" {
		return &statement, nil
	}

	return decodeDSSEEnvelope(data)
}

func decodeDSSEEnvelope(data []byte) (*InTotoStatement, error) {
	var envelope DSSEEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unable to parse DSSE envelope: %w", err)
	}

	if envelope.PayloadType == "" || envelope.Payload == "" {
		return nil, fmt.Errorf("invalid DSSE envelope: missing payload")
	}

	payloadBytes, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		payloadBytes, err = base64.URLEncoding.DecodeString(envelope.Payload)
		if err != nil {
			payloadBytes, err = base64.RawStdEncoding.DecodeString(envelope.Payload)
			if err != nil {
				return nil, fmt.Errorf("unable to decode payload: %w", err)
			}
		}
	}

	var statement InTotoStatement
	if err := json.Unmarshal(payloadBytes, &statement); err != nil {
		return nil, fmt.Errorf("unable to parse in-toto statement: %w", err)
	}

	return &statement, nil
}

func digestFromString(s string) digest.Digest {
	return digest.Digest(s)
}

func fetchReferrers(ctx context.Context, repo *remote.Repository, desc v1.Descriptor) ([]v1.Descriptor, error) {
	var referrers []v1.Descriptor

	err := repo.Referrers(ctx, desc, "", func(refs []v1.Descriptor) error {
		referrers = append(referrers, refs...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return referrers, nil
}

func isSBOMArtifact(desc v1.Descriptor) bool {
	at := desc.ArtifactType
	mt := desc.MediaType
	return strings.Contains(at, "sbom") ||
		strings.Contains(mt, "sbom") ||
		strings.Contains(at, "spdx") ||
		strings.Contains(mt, "spdx") ||
		strings.Contains(at, "cyclonedx") ||
		strings.Contains(mt, "cyclonedx")
}

func isSignatureArtifact(desc v1.Descriptor) bool {
	at := desc.ArtifactType
	mt := desc.MediaType
	return strings.Contains(at, "sig") ||
		strings.Contains(mt, "sig") ||
		strings.Contains(at, "cosign") ||
		strings.Contains(mt, "cosign") ||
		strings.Contains(at, "notation") ||
		strings.Contains(mt, "notation")
}

func isAttestationArtifact(desc v1.Descriptor) bool {
	at := desc.ArtifactType
	return strings.Contains(at, "attestation") ||
		strings.Contains(at, "intoto") ||
		strings.Contains(at, "in-toto")
}

func detectSBOMFormat(artifactType, mediaType string) string {
	combined := artifactType + mediaType
	if strings.Contains(combined, "spdx") {
		return "spdx"
	}
	if strings.Contains(combined, "cyclonedx") {
		return "cyclonedx"
	}
	return "unknown"
}

func detectSBOMFromAttestationManifests(manifests []AttestationManifest) *SBOM {
	for _, manifest := range manifests {
		for _, layer := range manifest.Layers {
			if format := detectSBOMPredicateType(layer.PredicateType); format != "" {
				return &SBOM{
					Present: true,
					Format:  format,
				}
			}
		}
	}
	return nil
}

func detectSBOMPredicateType(predicateType string) string {
	switch {
	case strings.Contains(predicateType, "spdx.dev"):
		return "spdx"
	case strings.Contains(predicateType, "cyclonedx.org"):
		return "cyclonedx"
	default:
		return ""
	}
}

type CosignTagResult struct {
	Signed       bool
	Signatures   []Signature
	Attestations []Attestation
	SBOM         *SBOM
}

func fetchCosignTagBasedArtifacts(
	ctx context.Context,
	repo *remote.Repository,
	desc v1.Descriptor,
	guardrails *FetchGuardrails,
) (*CosignTagResult, error) {
	result := &CosignTagResult{}

	digestStr := desc.Digest.String()
	cosignTagBase := strings.Replace(digestStr, ":", "-", 1)

	sigTag := cosignTagBase + ".sig"
	hasSig, err := fetchCosignSignatureTag(ctx, repo, sigTag)
	if err != nil && (ctx.Err() != nil) {
		return nil, ctx.Err()
	}
	if hasSig {
		result.Signed = true
		result.Signatures = append(result.Signatures, Signature{
			Algorithm: MediaTypeCosignSignature,
		})
	}

	attTag := cosignTagBase + ".att"
	attestations, sbom, err := fetchCosignAttestationTag(ctx, repo, attTag, guardrails)
	if err != nil && (ctx.Err() != nil) {
		return nil, ctx.Err()
	}
	if err == nil {
		result.Attestations = attestations
		result.SBOM = sbom
	}

	return result, nil
}

func fetchCosignSignatureTag(ctx context.Context, repo *remote.Repository, tag string) (bool, error) {
	desc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return false, err
	}

	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return false, err
	}
	defer rc.Close()

	rawJSON, err := io.ReadAll(rc)
	if err != nil {
		return false, err
	}

	var manifest v1.Manifest
	if err := json.Unmarshal(rawJSON, &manifest); err != nil {
		return false, err
	}

	for _, layer := range manifest.Layers {
		if layer.MediaType == MediaTypeCosignSignature {
			return true, nil
		}
	}

	return false, nil
}

func fetchCosignAttestationTag(
	ctx context.Context,
	repo *remote.Repository,
	tag string,
	guardrails *FetchGuardrails,
) ([]Attestation, *SBOM, error) {
	desc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return nil, nil, err
	}

	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return nil, nil, err
	}
	defer rc.Close()

	rawJSON, err := io.ReadAll(rc)
	if err != nil {
		return nil, nil, err
	}

	var manifest v1.Manifest
	if err := json.Unmarshal(rawJSON, &manifest); err != nil {
		return nil, nil, err
	}

	var attestations []Attestation
	var sbom *SBOM

	for i, layer := range manifest.Layers {
		if i >= guardrails.MaxLayersPerManifest {
			break
		}

		if layer.MediaType != MediaTypeDSSEEnvelope && layer.MediaType != MediaTypeInToto {
			continue
		}

		predicateType := ""
		if pt, ok := layer.Annotations["predicateType"]; ok {
			predicateType = pt
		} else if pt, ok := layer.Annotations[AnnotationPredicateType]; ok {
			predicateType = pt
		}

		if predicateType == "" && layer.Size <= guardrails.MaxBlobBytes {
			if decoded, err := fetchAndDecodeDSSELayer(ctx, repo, layer, guardrails); err == nil && decoded != nil {
				predicateType = decoded.PredicateType
			}
		}

		attestations = append(attestations, Attestation{
			Type: predicateType,
		})

		if format := detectSBOMPredicateType(predicateType); format != "" && sbom == nil {
			sbom = &SBOM{
				Present: true,
				Format:  format,
			}
		}
	}

	return attestations, sbom, nil
}

func fetchAndDecodeDSSELayer(
	ctx context.Context,
	repo *remote.Repository,
	layer v1.Descriptor,
	guardrails *FetchGuardrails,
) (*InTotoStatement, error) {
	if layer.Size > guardrails.MaxBlobBytes {
		return nil, fmt.Errorf("layer too large: %d > %d", layer.Size, guardrails.MaxBlobBytes)
	}

	rc, err := repo.Fetch(ctx, layer)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	rawJSON, err := io.ReadAll(io.LimitReader(rc, guardrails.MaxBlobBytes))
	if err != nil {
		return nil, err
	}

	return decodeDSSEEnvelope(rawJSON)
}
