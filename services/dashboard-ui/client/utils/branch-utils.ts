import type { TAppBranch, TAppBranchConfig, TAppBranchRun } from '@/types'

export const latestBranchConfig = (
  branch: TAppBranch
): TAppBranchConfig | undefined => {
  if (!branch.configs?.length) return undefined
  return [...branch.configs].sort(
    (a, b) => (b.config_number || 0) - (a.config_number || 0)
  )[0]
}

/**
 * Section path to keep when the branch switcher swaps in another branch id.
 * Only single-segment section/tab routes survive the switch; deeper routes are
 * resource details (runs, components, builds, policies, syncs) scoped to the
 * current branch, so they fall back to the branch overview.
 */
export const branchSwitchSectionPath = (
  pathname: string,
  branchBasePath: string
): string => {
  if (!pathname.startsWith(branchBasePath)) return ''
  const remainder = pathname.slice(branchBasePath.length)
  if (!remainder || !remainder.startsWith('/')) return ''
  const segments = remainder.split('/').filter(Boolean)
  return segments.length === 1 ? `/${segments[0]}` : ''
}

export const manualRunPinLabel = (
  branchRun?: TAppBranchRun
): string | undefined => {
  if (!branchRun) return undefined
  const trigger = branchRun.metadata?.trigger ?? branchRun.event_type
  if (trigger !== 'manual') return undefined

  const meta = branchRun.metadata
  const prNumber = branchRun.pr_number ?? meta?.pr_number
  if (prNumber != null) return `PR #${prNumber}`
  if (meta?.tag) return meta.tag

  const sha = branchRun.head_sha || meta?.head_sha || meta?.git_ref
  if (!sha) return undefined
  return sha.slice(0, 7)
}
