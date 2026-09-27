import {
  githubCommitUrl,
  resolvePrLink,
} from '@/components/branches/shared/pr-link'
import type { TAppBranchRun } from '@/types'

export type TRunSource =
  | {
      kind: 'pull-request'
      number: number
      url?: string
      label?: string
      baseBranch?: string
    }
  | { kind: 'tag'; tag: string; url?: string }
  | { kind: 'manual' }
  | { kind: 'commit' }

const TAG_REF_PREFIX = 'refs/tags/'

export const githubTagUrl = (repoSlug?: string, tag?: string) =>
  repoSlug && tag
    ? `https://github.com/${repoSlug}/releases/tag/${encodeURIComponent(tag)}`
    : undefined

export const resolveRunSource = (
  branchRun?: TAppBranchRun,
  repoSlug?: string
): TRunSource => {
  const meta = branchRun?.metadata
  const gitRef = meta?.git_ref
  const tag =
    meta?.tag ||
    (gitRef?.startsWith(TAG_REF_PREFIX)
      ? gitRef.slice(TAG_REF_PREFIX.length)
      : undefined)

  if (meta?.trigger === 'tag' || tag) {
    if (tag) return { kind: 'tag', tag, url: githubTagUrl(repoSlug, tag) }
  }

  if (meta?.trigger === 'manual' || branchRun?.run_type === 'manual-run') {
    return { kind: 'manual' }
  }

  const pr = resolvePrLink({
    repoSlug,
    prNumber: branchRun?.pr_number ?? meta?.pr_number,
    commitMessage: branchRun?.vcs_connection_commit?.message,
  })
  const number = pr?.number ?? branchRun?.pr_number ?? meta?.pr_number
  if (number != null) {
    return {
      kind: 'pull-request',
      number,
      url: pr?.url,
      label: meta?.trigger === 'github_label' ? meta.github_label : undefined,
      baseBranch: branchRun?.base_branch || meta?.base_branch,
    }
  }

  return { kind: 'commit' }
}

export const commitUrl = (repoSlug?: string, sha?: string) =>
  githubCommitUrl(repoSlug, sha)
