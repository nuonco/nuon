export type TGitReferenceTarget =
  | { type: 'commit'; sha: string }
  | { type: 'pull-request'; number: number }
  | { type: 'tag'; tag: string }

type TVcsHost = 'github' | 'gitlab' | 'bitbucket'

const VCS_HOSTS: { host: TVcsHost; marker: string }[] = [
  { host: 'github', marker: 'github.com' },
  { host: 'gitlab', marker: 'gitlab.com' },
  { host: 'bitbucket', marker: 'bitbucket.org' },
]

export const externalGitUrl = (href?: string) => {
  if (!href) return undefined
  const trimmed = href.trim()
  if (/^https?:\/\//i.test(trimmed)) return trimmed
  if (/^(?:github\.com|gitlab\.com|bitbucket\.org)\//i.test(trimmed)) {
    return `https://${trimmed}`
  }
  return undefined
}

const repoWebBase = (repo: string) => {
  const clean = repo.trim().replace(/\.git$/i, '').replace(/\/+$/, '')
  if (!clean || /\s/.test(clean)) return undefined

  const absolute = /^https?:\/\//i.test(clean)
    ? clean
    : /^(?:github\.com|gitlab\.com|bitbucket\.org)\//i.test(clean)
      ? `https://${clean}`
      : undefined

  if (absolute) {
    const match = VCS_HOSTS.find((item) => absolute.includes(item.marker))
    if (!match) return undefined
    return { host: match.host, base: absolute }
  }

  if (clean.includes('://')) return undefined
  return { host: 'github' as const, base: `https://github.com/${clean}` }
}

export const gitReferenceUrl = (
  repo: string | undefined,
  target: TGitReferenceTarget
) => {
  if (!repo) return undefined
  const parsed = repoWebBase(repo)
  if (!parsed) return undefined
  const { host, base } = parsed

  if (target.type === 'commit') {
    if (!target.sha) return undefined
    if (host === 'gitlab') return `${base}/-/commit/${target.sha}`
    if (host === 'bitbucket') return `${base}/commits/${target.sha}`
    return `${base}/commit/${target.sha}`
  }

  if (target.type === 'pull-request') {
    if (!Number.isInteger(target.number) || target.number < 1) return undefined
    if (host === 'gitlab') return `${base}/-/merge_requests/${target.number}`
    if (host === 'bitbucket') return `${base}/pull-requests/${target.number}`
    return `${base}/pull/${target.number}`
  }

  if (!target.tag) return undefined
  const tag = encodeURIComponent(target.tag)
  if (host === 'gitlab') return `${base}/-/tags/${tag}`
  if (host === 'bitbucket') return `${base}/src/${tag}`
  return `${base}/releases/tag/${tag}`
}

export const buildCommitUrl = (
  repo: string | undefined,
  sha: string | undefined
): string | null =>
  sha ? (gitReferenceUrl(repo, { type: 'commit', sha }) ?? null) : null

export const vcsRepo = (
  config?: {
    connected_github_vcs_config?: { repo?: string } | null
    public_git_vcs_config?: { repo?: string } | null
  } | null
) =>
  config?.connected_github_vcs_config?.repo ||
  config?.public_git_vcs_config?.repo ||
  undefined
