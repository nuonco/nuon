// Package imageref renders machine-readable and human-friendly references
// for image-type component builds.
//
// All builds for image components (Docker builds, external images) populate a
// common set of source-identity fields on the build record:
//
//   - SourceImage  — repo only, e.g. "nginx"
//   - SourceRef    — what the user wrote, e.g. "nginx:1.25.3"
//   - ResolvedTag  — the tag the runner actually pulled, e.g. "1.25.5"
//   - SourceDigest — manifest list digest, e.g. "sha256:abc..."
//
// This package converts those fields into the canonical forms used across
// the platform:
//
//   - [ImageRef]   for pod specs, Helm values, and rendered K8s manifests
//     (always digest-only — `repo@sha256:...`)
//   - [DisplayRef] for Nuon plan output, dashboard, CLI, and audit logs
//     (human-friendly — `repo:tag (sha256:short)`)
//
// The package takes a primitive [Source] struct so that both ctl-api
// (which uses internal app models) and the CLI (which uses generated SDK
// models) can call it without importing each other.
package imageref

import (
	"fmt"
	"strings"
)

// shortDigestLen is the number of hex characters of the digest payload
// included in display-form references. Matches the convention used by Docker
// and oras (12 chars by Docker, 7 by oras/git short SHAs — we use 7 to keep
// the display compact and consistent with git short SHAs).
const shortDigestLen = 7

// Source captures the image-source identity fields needed to render
// references for an image-type component build. ctl-api and the CLI populate
// this struct from their own ComponentBuild model types.
type Source struct {
	// SourceImage is the bare repository portion of the user's spec,
	// e.g. "nginx" or "ghcr.io/nuonco/foo". Always set for image builds.
	SourceImage string
	// SourceRef is the original reference the user wrote, e.g.
	// "nginx:1.25.3" or "nginx@sha256:abc...". Always set for image builds.
	SourceRef string
	// ResolvedTag is the tag the runner actually pulled, when known.
	// Empty for digest-pinned references (where the user wrote
	// "repo@sha256:...").
	ResolvedTag string
	// SourceDigest is the manifest list digest of the resolved image.
	// Empty for legacy builds without recorded source identity.
	SourceDigest string
}

// ImageRef returns the canonical machine-readable reference for use in
// Kubernetes pod specs, Helm values, and rendered manifests. The reference
// is always digest-only:
//
//	nginx@sha256:abc...
//
// Returns "" when the build has no SourceDigest. Callers should fall back to
// legacy behavior when the result is empty.
func ImageRef(s Source) string {
	if s.SourceDigest == "" || s.SourceImage == "" {
		return ""
	}
	return fmt.Sprintf("%s@%s", s.SourceImage, s.SourceDigest)
}

// DisplayRef returns a human-friendly reference for plan output, the
// dashboard, the CLI, and audit logs. The form is:
//
//	repo:tag (sha256:abcdef0)
//
// Tag preference is ResolvedTag → tag parsed from SourceRef → none. When
// no tag is available the result falls back to a digest-only display:
//
//	repo@sha256:abcdef0
//
// When both digest and tag are unavailable, falls back to whichever of
// SourceRef or SourceImage is set (keeps the function safe to call on
// legacy builds without surfacing an empty string).
func DisplayRef(s Source) string {
	tag := s.ResolvedTag
	if tag == "" {
		tag = parseTagFromRef(s.SourceRef)
	}

	short := shortDigest(s.SourceDigest)

	switch {
	case s.SourceImage != "" && tag != "" && short != "":
		return fmt.Sprintf("%s:%s (sha256:%s)", s.SourceImage, tag, short)
	case s.SourceImage != "" && tag != "":
		return fmt.Sprintf("%s:%s", s.SourceImage, tag)
	case s.SourceImage != "" && short != "":
		return fmt.Sprintf("%s@sha256:%s", s.SourceImage, short)
	case s.SourceRef != "":
		return s.SourceRef
	default:
		return s.SourceImage
	}
}

// parseTagFromRef extracts the tag portion from a user-supplied reference.
// Returns "" for digest-pinned references and references with no explicit
// tag.
//
// Examples:
//
//	"nginx:1.25.3"           → "1.25.3"
//	"nginx@sha256:abc..."    → ""
//	"nginx"                  → ""
//	"ghcr.io/foo/bar:1.0"    → "1.0"
//	"ghcr.io:5000/foo:1.0"   → "1.0"
func parseTagFromRef(ref string) string {
	if ref == "" {
		return ""
	}
	// Strip any digest suffix; we want the tag, not the digest.
	if at := strings.Index(ref, "@"); at != -1 {
		ref = ref[:at]
	}
	// The tag is everything after the LAST colon, but only if that colon
	// comes after the last "/" (otherwise it's a registry port like
	// "ghcr.io:5000").
	colon := strings.LastIndex(ref, ":")
	if colon == -1 {
		return ""
	}
	if slash := strings.LastIndex(ref, "/"); slash != -1 && slash > colon {
		return ""
	}
	return ref[colon+1:]
}

// shortDigest returns the first shortDigestLen hex chars of a "sha256:..."
// digest, with no algorithm prefix. Returns "" when the input is not a
// recognisable sha256 digest.
func shortDigest(digest string) string {
	const prefix = "sha256:"
	if !strings.HasPrefix(digest, prefix) {
		return ""
	}
	hex := digest[len(prefix):]
	if len(hex) < shortDigestLen {
		return hex
	}
	return hex[:shortDigestLen]
}
