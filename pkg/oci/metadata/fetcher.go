package metadata

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	SBOMMediaTypeSPDX      = "application/spdx+json"
	SBOMMediaTypeCycloneDX = "application/vnd.cyclonedx+json"
	SignatureMediaType     = "application/vnd.dev.cosign.simplesigning.v1+json"

	ArtifactTypeSBOM      = "application/vnd.oci.artifact.sbom.v1+json"
	ArtifactTypeSignature = "application/vnd.dev.cosign.artifact.sig.v1+json"

	// OCI image index media types
	MediaTypeImageIndex     = "application/vnd.oci.image.index.v1+json"
	MediaTypeDockerManifest = "application/vnd.docker.distribution.manifest.list.v2+json"

	// Attestation-related annotations and media types
	AnnotationReferenceType   = "vnd.docker.reference.type"
	AnnotationReferenceDigest = "vnd.docker.reference.digest"
	AnnotationPredicateType   = "in-toto.io/predicate-type"
	ReferenceTypeAttestation  = "attestation-manifest"

	MediaTypeInToto = "application/vnd.in-toto+json"
)

type RegistryAuth struct {
	ServerAddress string
	Username      string
	Password      string
}

// FetchGuardrails defines limits for fetching attestation content.
type FetchGuardrails struct {
	MaxBlobBytes         int64
	MaxTotalBytes        int64
	MaxAttestations      int
	MaxLayersPerManifest int
}

// DefaultGuardrails returns sensible default limits for attestation fetching.
func DefaultGuardrails() FetchGuardrails {
	return FetchGuardrails{
		MaxBlobBytes:         1 * 1024 * 1024,  // 1MB per blob
		MaxTotalBytes:        10 * 1024 * 1024, // 10MB total
		MaxAttestations:      10,
		MaxLayersPerManifest: 5,
	}
}

type FetchOptions struct {
	Image  string
	Tag    string
	Auth   *RegistryAuth
	Digest string

	// Layer fetch controls
	IncludeIndex                bool
	IncludeAttestationManifests bool
	IncludeAttestationLayers    bool

	// Platform filter (e.g., "linux/amd64")
	Platform string

	// Guardrails for limiting fetch sizes
	Guardrails *FetchGuardrails
}

func FetchImageMetadata(ctx context.Context, opts *FetchOptions) (*ImageMetadata, error) {
	repo, err := remote.NewRepository(opts.Image)
	if err != nil {
		return nil, fmt.Errorf("unable to create repository client: %w", err)
	}

	if opts.Auth != nil && opts.Auth.Username != "" {
		serverAddr := opts.Auth.ServerAddress
		if serverAddr == "" {
			parts := strings.SplitN(opts.Image, "/", 2)
			if len(parts) > 0 {
				serverAddr = parts[0]
			}
		}
		repo.Client = &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.DefaultCache,
			Credential: auth.StaticCredential(serverAddr, auth.Credential{
				Username: opts.Auth.Username,
				Password: opts.Auth.Password,
			}),
		}
	}

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

	// Fetch Layer 1: Image Index if requested
	if opts.IncludeIndex || opts.IncludeAttestationManifests {
		index, err := fetchImageIndex(ctx, repo, desc)
		if err != nil {
			if !errors.Is(err, errdef.ErrNotFound) && !isNotIndexError(err) {
				return nil, fmt.Errorf("unable to fetch image index: %w", err)
			}
		} else {
			result.Index = index

			// Fetch Layer 2: Attestation Manifests if requested
			if opts.IncludeAttestationManifests && index != nil {
				attestationManifests, err := fetchAttestationManifests(ctx, repo, index, opts, guardrails)
				if err != nil {
					return nil, fmt.Errorf("unable to fetch attestation manifests: %w", err)
				}
				result.AttestationManifests = attestationManifests
			}
		}
	}

	// Continue with referrers-based metadata (signatures, SBOMs, etc.)
	referrers, err := fetchReferrers(ctx, repo, desc)
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return result, nil
		}
		return nil, fmt.Errorf("unable to fetch referrers: %w", err)
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

	return result, nil
}

func isNotIndexError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not an image index")
}

// fetchImageIndex fetches and parses the image index (manifest list).
func fetchImageIndex(ctx context.Context, repo *remote.Repository, desc v1.Descriptor) (*ImageIndex, error) {
	// Check if this is an index media type
	if !isIndexMediaType(desc.MediaType) {
		return nil, fmt.Errorf("not an image index: media type is %s", desc.MediaType)
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
			Annotations: m.Annotations,
		}

		if m.Platform != nil {
			entry.Platform = &Platform{
				OS:           m.Platform.OS,
				Architecture: m.Platform.Architecture,
				Variant:      m.Platform.Variant,
			}
		}

		// Check if this is an attestation manifest
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

// fetchAttestationManifests fetches attestation manifests from the index.
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

		// Apply platform filter if specified
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

// fetchAttestationManifest fetches a single attestation manifest and optionally its layers.
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

	// Extract reference digest from annotations
	if refDigest, ok := entry.Annotations[AnnotationReferenceDigest]; ok {
		manifest.RefDigest = refDigest
	}

	// Fetch Layer 3: Attestation Layers if requested
	if opts.IncludeAttestationLayers {
		layers, layerBytes, err := fetchAttestationLayers(ctx, repo, ociManifest.Layers, guardrails, currentTotalBytes+bytesRead)
		if err != nil {
			return nil, bytesRead, fmt.Errorf("unable to fetch attestation layers: %w", err)
		}
		manifest.Layers = layers
		bytesRead += layerBytes
	} else {
		// Just extract layer metadata without fetching content
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

// fetchAttestationLayers fetches and decodes attestation layer blobs.
func fetchAttestationLayers(
	ctx context.Context,
	repo *remote.Repository,
	layers []v1.Descriptor,
	guardrails *FetchGuardrails,
	currentTotalBytes int64,
) ([]AttestationLayer, int64, error) {
	var result []AttestationLayer
	var bytesRead int64

	for i, l := range layers {
		if i >= guardrails.MaxLayersPerManifest {
			break
		}

		if currentTotalBytes+bytesRead+l.Size > guardrails.MaxTotalBytes {
			layer := AttestationLayer{
				Digest:    l.Digest.String(),
				MediaType: l.MediaType,
				Size:      l.Size,
				Truncated: true,
			}
			if predicateType, ok := l.Annotations[AnnotationPredicateType]; ok {
				layer.PredicateType = predicateType
			}
			result = append(result, layer)
			continue
		}

		layer, layerBytes, err := fetchAttestationLayer(ctx, repo, l, guardrails)
		if err != nil {
			return nil, bytesRead, fmt.Errorf("unable to fetch layer %s: %w", l.Digest.String(), err)
		}

		bytesRead += layerBytes
		result = append(result, *layer)
	}

	return result, bytesRead, nil
}

// fetchAttestationLayer fetches a single attestation layer and decodes its content.
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

	// Check if blob is too large
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

	// Try to decode as DSSE envelope
	decoded, err := decodeDSSEEnvelope(rawJSON)
	if err == nil && decoded != nil {
		layer.Decoded = decoded
		if layer.PredicateType == "" && decoded.PredicateType != "" {
			layer.PredicateType = decoded.PredicateType
		}
	}

	return layer, bytesRead, nil
}

// decodeDSSEEnvelope decodes a DSSE envelope and extracts the in-toto statement.
func decodeDSSEEnvelope(data []byte) (*InTotoStatement, error) {
	var envelope DSSEEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("unable to parse DSSE envelope: %w", err)
	}

	if envelope.PayloadType == "" || envelope.Payload == "" {
		return nil, fmt.Errorf("invalid DSSE envelope: missing payload")
	}

	// Decode base64 payload
	payloadBytes, err := base64.StdEncoding.DecodeString(envelope.Payload)
	if err != nil {
		// Try URL-safe base64
		payloadBytes, err = base64.URLEncoding.DecodeString(envelope.Payload)
		if err != nil {
			// Try raw/unpadded base64
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
