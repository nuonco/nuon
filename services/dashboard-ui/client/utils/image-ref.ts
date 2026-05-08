/**
 * Image-reference helpers for image-type component builds.
 *
 * Mirrors the Go helpers in `pkg/oci/imageref/imageref.go` so the
 * dashboard renders the same human-friendly form for image refs as the
 * CLI and audit logs.
 *
 * - `imageRef(b)`     → "repo@sha256:..." (digest-only, machine ref).
 *                       Empty for legacy builds without source_digest.
 * - `displayRef(b)`   → "repo:tag (sha256:abcdef0)" (human-friendly,
 *                       falls back gracefully for partial info).
 */
export type TImageBuildSource = {
  source_image?: string
  source_ref?: string
  resolved_tag?: string
  source_digest?: string
}

const SHORT_DIGEST_LEN = 7

const parseTagFromRef = (ref?: string): string => {
  if (!ref) return ''
  const at = ref.indexOf('@')
  const stripped = at !== -1 ? ref.slice(0, at) : ref
  const colon = stripped.lastIndexOf(':')
  if (colon === -1) return ''
  const slash = stripped.lastIndexOf('/')
  if (slash !== -1 && slash > colon) return ''
  return stripped.slice(colon + 1)
}

const shortDigest = (digest?: string): string => {
  const prefix = 'sha256:'
  if (!digest || !digest.startsWith(prefix)) return ''
  const hex = digest.slice(prefix.length)
  return hex.slice(0, SHORT_DIGEST_LEN)
}

export const imageRef = (s: TImageBuildSource): string => {
  if (!s.source_digest || !s.source_image) return ''
  return `${s.source_image}@${s.source_digest}`
}

export const displayRef = (s: TImageBuildSource): string => {
  const tag = s.resolved_tag || parseTagFromRef(s.source_ref)
  const short = shortDigest(s.source_digest)

  if (s.source_image && tag && short) {
    return `${s.source_image}:${tag} (sha256:${short})`
  }
  if (s.source_image && tag) {
    return `${s.source_image}:${tag}`
  }
  if (s.source_image && short) {
    return `${s.source_image}@sha256:${short}`
  }
  return s.source_ref || s.source_image || ''
}

export const isImageBuild = (s: TImageBuildSource): boolean =>
  Boolean(s.source_ref || s.source_image || s.source_digest)
