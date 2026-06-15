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

// Spec is the raw image source a build planner hands the runner: the
// image_url, the literal tag (which may itself be a bare "sha256:..." digest
// or a full "repo@sha256:..." reference), and an optional update_policy semver
// constraint. The same `@sha256:` shapes are parsed once here so the pull
// reference and the recorded [Source] identity never disagree.
type Spec struct {
	// Image is the image_url. It may carry a "@sha256:..." digest when the
	// user pinned by digest in the image field rather than the tag.
	Image string
	// Tag is the literal tag. It may be a plain tag ("1.25.3"), a bare
	// digest ("sha256:..."), or a full "repo@sha256:..." reference.
	Tag string
	// UpdatePolicy is the semver constraint. When set it takes precedence:
	// the runner selects a concrete tag at build time and passes it back as
	// selectedTag.
	UpdatePolicy string
}

// parseSpec decomposes a planner-provided image+tag into its bare repository,
// literal tag, and digest. At most one of tag/digest is set. A "@sha256:..."
// substring in either field marks a digest-pinned reference.
func parseSpec(image, tag string) (repo, literalTag, digest string) {
	repo = image
	switch {
	case strings.Contains(tag, "@sha256:"):
		// User wrote "repo@sha256:..." and the planner kept it in the tag.
		parts := strings.SplitN(tag, "@sha256:", 2)
		if image == "" {
			repo = parts[0]
		}
		digest = "sha256:" + parts[1]
	case strings.HasPrefix(tag, "sha256:"):
		digest = tag
	case tag != "":
		literalTag = tag
	case strings.Contains(image, "@sha256:"):
		// Digest baked into the image field; tag (if any) is ignored.
		parts := strings.SplitN(image, "@sha256:", 2)
		repo = parts[0]
		digest = "sha256:" + parts[1]
	}
	return repo, literalTag, digest
}

// PullRef returns the repo-relative reference to hand the OCI resolver and
// copier: the update_policy-selected tag when a policy is set, otherwise a
// plain tag or a bare "sha256:..." digest. The source repository config is
// already scoped to the repository, so any "repo@" prefix is stripped to the
// bare digest. Returns "" only when the spec carries no tag and no digest, in
// which case the resolver surfaces a clear error.
func (sp Spec) PullRef(selectedTag string) string {
	if sp.UpdatePolicy != "" {
		return selectedTag
	}
	_, literalTag, digest := parseSpec(sp.Image, sp.Tag)
	if digest != "" {
		return digest
	}
	return literalTag
}

// Identity returns the source-identity fields to record on the ComponentBuild
// row. SourceDigest is left empty: the caller fills it in after resolving the
// manifest. selectedTag is the update_policy-selected tag (ignored when no
// policy is set).
//
//   - SourceRef is what the user asked for: "repo:tag" for tag-based refs,
//     "repo@sha256:..." for digest-pinned refs, or "repo:<update_policy>" when
//     a semver constraint is set (the constraint, not the tag we happened to
//     select on this build).
//   - SourceImage is the bare repository.
//   - ResolvedTag is the tag the runner pulled: the selected tag for
//     update_policy, the literal tag for tag-based refs, empty for digest pins.
func (sp Spec) Identity(selectedTag string) Source {
	repo, literalTag, digest := parseSpec(sp.Image, sp.Tag)
	s := Source{SourceImage: repo}
	switch {
	case sp.UpdatePolicy != "":
		s.SourceRef = fmt.Sprintf("%s:%s", repo, sp.UpdatePolicy)
		s.ResolvedTag = selectedTag
	case digest != "":
		s.SourceRef = fmt.Sprintf("%s@%s", repo, digest)
	default:
		s.SourceRef = fmt.Sprintf("%s:%s", repo, literalTag)
		s.ResolvedTag = literalTag
	}
	return s
}

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
