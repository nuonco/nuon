package metadata

import "encoding/json"

type ImageMetadata struct {
	Image        string        `json:"image" temporaljson:"image"`
	Tag          string        `json:"tag" temporaljson:"tag"`
	Digest       string        `json:"digest" temporaljson:"digest"`
	SBOM         *SBOM         `json:"sbom,omitempty" temporaljson:"sbom,omitempty"`
	Signatures   []Signature   `json:"signatures,omitempty" temporaljson:"signatures,omitempty"`
	Attestations []Attestation `json:"attestations,omitempty" temporaljson:"attestations,omitempty"`
	Signed       bool          `json:"signed" temporaljson:"signed"`

	// Layer 1: Image Index (manifest list)
	Index *ImageIndex `json:"index,omitempty" temporaljson:"index,omitempty"`

	// Layer 2: Attestation Manifests
	AttestationManifests []AttestationManifest `json:"attestation_manifests,omitempty" temporaljson:"attestation_manifests,omitempty"`
}

type SBOM struct {
	Present bool   `json:"present" temporaljson:"present"`
	Format  string `json:"format,omitempty" temporaljson:"format,omitempty"`
	URI     string `json:"uri,omitempty" temporaljson:"uri,omitempty"`
}

type Signature struct {
	KeyID     string `json:"key_id,omitempty" temporaljson:"key_id,omitempty"`
	Issuer    string `json:"issuer,omitempty" temporaljson:"issuer,omitempty"`
	Subject   string `json:"subject,omitempty" temporaljson:"subject,omitempty"`
	Algorithm string `json:"algorithm,omitempty" temporaljson:"algorithm,omitempty"`
}

type Attestation struct {
	Type      string `json:"type" temporaljson:"type"`
	Predicate string `json:"predicate,omitempty" temporaljson:"predicate,omitempty"`
}

// Platform represents an OCI platform specification.
type Platform struct {
	OS           string `json:"os" temporaljson:"os"`
	Architecture string `json:"architecture" temporaljson:"architecture"`
	Variant      string `json:"variant,omitempty" temporaljson:"variant,omitempty"`
}

// ManifestEntry represents a manifest within an image index.
type ManifestEntry struct {
	Digest        string            `json:"digest" temporaljson:"digest"`
	MediaType     string            `json:"media_type" temporaljson:"media_type"`
	Size          int64             `json:"size" temporaljson:"size"`
	Platform      *Platform         `json:"platform,omitempty" temporaljson:"platform,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty" temporaljson:"annotations,omitempty"`
	IsAttestation bool              `json:"is_attestation" temporaljson:"is_attestation"`
}

// ImageIndex represents Layer 1 - the image index (manifest list).
type ImageIndex struct {
	Digest    string          `json:"digest" temporaljson:"digest"`
	MediaType string          `json:"media_type" temporaljson:"media_type"`
	RawJSON   json.RawMessage `json:"raw_json,omitempty" temporaljson:"raw_json,omitempty"`
	Manifests []ManifestEntry `json:"manifests" temporaljson:"manifests"`
}

// AttestationManifest represents Layer 2 - an attestation manifest for a specific platform.
type AttestationManifest struct {
	Digest      string             `json:"digest" temporaljson:"digest"`
	MediaType   string             `json:"media_type" temporaljson:"media_type"`
	Platform    *Platform          `json:"platform,omitempty" temporaljson:"platform,omitempty"`
	RefDigest   string             `json:"ref_digest,omitempty" temporaljson:"ref_digest,omitempty"`
	Annotations map[string]string  `json:"annotations,omitempty" temporaljson:"annotations,omitempty"`
	RawJSON     json.RawMessage    `json:"raw_json,omitempty" temporaljson:"raw_json,omitempty"`
	Layers      []AttestationLayer `json:"layers,omitempty" temporaljson:"layers,omitempty"`
}

// AttestationLayer represents Layer 3 - an attestation blob containing DSSE/in-toto content.
type AttestationLayer struct {
	Digest        string           `json:"digest" temporaljson:"digest"`
	MediaType     string           `json:"media_type" temporaljson:"media_type"`
	Size          int64            `json:"size" temporaljson:"size"`
	PredicateType string           `json:"predicate_type,omitempty" temporaljson:"predicate_type,omitempty"`
	RawJSON       json.RawMessage  `json:"raw_json,omitempty" temporaljson:"raw_json,omitempty"`
	Decoded       *InTotoStatement `json:"decoded,omitempty" temporaljson:"decoded,omitempty"`
	Truncated     bool             `json:"truncated,omitempty" temporaljson:"truncated,omitempty"`
}

// DSSEEnvelope represents a Dead Simple Signing Envelope.
type DSSEEnvelope struct {
	PayloadType string          `json:"payloadType" temporaljson:"payloadType"`
	Payload     string          `json:"payload" temporaljson:"payload"`
	Signatures  []DSSESignature `json:"signatures,omitempty" temporaljson:"signatures,omitempty"`
}

// DSSESignature represents a signature in a DSSE envelope.
type DSSESignature struct {
	KeyID string `json:"keyid,omitempty" temporaljson:"keyid,omitempty"`
	Sig   string `json:"sig" temporaljson:"sig"`
}

// InTotoStatement represents an in-toto statement from the attestation.
type InTotoStatement struct {
	Type          string          `json:"_type" temporaljson:"_type"`
	Subject       []InTotoSubject `json:"subject,omitempty" temporaljson:"subject,omitempty"`
	PredicateType string          `json:"predicateType" temporaljson:"predicateType"`
	Predicate     json.RawMessage `json:"predicate,omitempty" temporaljson:"predicate,omitempty"`
}

// InTotoSubject represents a subject in an in-toto statement.
type InTotoSubject struct {
	Name   string            `json:"name" temporaljson:"name"`
	Digest map[string]string `json:"digest,omitempty" temporaljson:"digest,omitempty"`
}

type ExternalImagePolicyInput struct {
	Image    string         `json:"image" temporaljson:"image"`
	Tag      string         `json:"tag" temporaljson:"tag"`
	Digest   string         `json:"digest" temporaljson:"digest"`
	Metadata *ImageMetadata `json:"metadata" temporaljson:"metadata"`
}
