package metadata

import "encoding/json"

type ImageMetadata struct {
	Image        string        `json:"image"`
	Tag          string        `json:"tag"`
	Digest       string        `json:"digest"`
	SBOM         *SBOM         `json:"sbom,omitempty"`
	Signatures   []Signature   `json:"signatures,omitempty"`
	Attestations []Attestation `json:"attestations,omitempty"`
	Signed       bool          `json:"signed"`

	// Layer 1: Image Index (manifest list)
	Index *ImageIndex `json:"index,omitempty"`

	// Layer 2: Attestation Manifests
	AttestationManifests []AttestationManifest `json:"attestation_manifests,omitempty"`
}

type SBOM struct {
	Present bool   `json:"present"`
	Format  string `json:"format,omitempty"`
	URI     string `json:"uri,omitempty"`
}

type Signature struct {
	KeyID     string `json:"key_id,omitempty"`
	Issuer    string `json:"issuer,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
}

type Attestation struct {
	Type      string `json:"type"`
	Predicate string `json:"predicate,omitempty"`
}

// Platform represents an OCI platform specification.
type Platform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	Variant      string `json:"variant,omitempty"`
}

// ManifestEntry represents a manifest within an image index.
type ManifestEntry struct {
	Digest        string            `json:"digest"`
	MediaType     string            `json:"media_type"`
	Platform      *Platform         `json:"platform,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	IsAttestation bool              `json:"is_attestation"`
}

// ImageIndex represents Layer 1 - the image index (manifest list).
type ImageIndex struct {
	Digest    string          `json:"digest"`
	MediaType string          `json:"media_type"`
	RawJSON   json.RawMessage `json:"raw_json,omitempty"`
	Manifests []ManifestEntry `json:"manifests"`
}

// AttestationManifest represents Layer 2 - an attestation manifest for a specific platform.
type AttestationManifest struct {
	Digest      string             `json:"digest"`
	MediaType   string             `json:"media_type"`
	Platform    *Platform          `json:"platform,omitempty"`
	RefDigest   string             `json:"ref_digest,omitempty"`
	Annotations map[string]string  `json:"annotations,omitempty"`
	RawJSON     json.RawMessage    `json:"raw_json,omitempty"`
	Layers      []AttestationLayer `json:"layers,omitempty"`
}

// AttestationLayer represents Layer 3 - an attestation blob containing DSSE/in-toto content.
type AttestationLayer struct {
	Digest        string           `json:"digest"`
	MediaType     string           `json:"media_type"`
	Size          int64            `json:"size"`
	PredicateType string           `json:"predicate_type,omitempty"`
	RawJSON       json.RawMessage  `json:"raw_json,omitempty"`
	Decoded       *InTotoStatement `json:"decoded,omitempty"`
	Truncated     bool             `json:"truncated,omitempty"`
}

// DSSEEnvelope represents a Dead Simple Signing Envelope.
type DSSEEnvelope struct {
	PayloadType string          `json:"payloadType"`
	Payload     string          `json:"payload"`
	Signatures  []DSSESignature `json:"signatures,omitempty"`
}

// DSSESignature represents a signature in a DSSE envelope.
type DSSESignature struct {
	KeyID string `json:"keyid,omitempty"`
	Sig   string `json:"sig"`
}

// InTotoStatement represents an in-toto statement from the attestation.
type InTotoStatement struct {
	Type          string          `json:"_type"`
	Subject       []InTotoSubject `json:"subject,omitempty"`
	PredicateType string          `json:"predicateType"`
	Predicate     json.RawMessage `json:"predicate,omitempty"`
}

// InTotoSubject represents a subject in an in-toto statement.
type InTotoSubject struct {
	Name   string            `json:"name"`
	Digest map[string]string `json:"digest,omitempty"`
}

type ExternalImagePolicyInput struct {
	Image    string         `json:"image"`
	Tag      string         `json:"tag"`
	Digest   string         `json:"digest"`
	Metadata *ImageMetadata `json:"metadata"`
}
