import { getStepDisplayStatus } from '@/components/branches/shared/step-status'
import { isPreviewBranchRun } from '@/components/branches/shared/preview-run-utils'
import type { TInstallWorkflow } from '@/types'
import { commitUrl, resolveRunSource, type TRunSource } from './run-source'

const PLAN_GROUP_STEP = /^plan install group:\s*(.+)$/i

export type TGroupApprovedVersion = {
  configId?: string
  version?: number
  source: TRunSource
  sha?: string
  shaUrl?: string
  message?: string
  author?: string
  createdAt?: string
}

const groupNameKey = (name: string) => name.trim().toLowerCase()

// Newest approved plan wins. Runs are expected newest first.
export const lastApprovedGroupVersions = (
  runs: TInstallWorkflow[],
  repoSlug?: string
): Record<string, TGroupApprovedVersion> => {
  const found: Record<string, TGroupApprovedVersion> = {}

  for (const run of runs) {
    const branchRun = run.app_branch_runs?.[0]
    if (!branchRun || isPreviewBranchRun(branchRun) || run.plan_only) continue

    for (const step of run.steps ?? []) {
      const match = step.name?.match(PLAN_GROUP_STEP)
      if (!match) continue
      const name = groupNameKey(match[1])
      if (!name || found[name]) continue
      if (getStepDisplayStatus(step) !== 'approved') continue

      const sha =
        branchRun.vcs_connection_commit?.sha ||
        branchRun.head_sha ||
        branchRun.metadata?.head_sha
      found[name] = {
        configId: branchRun.app_config_id,
        source: resolveRunSource(branchRun, repoSlug),
        sha,
        shaUrl: commitUrl(repoSlug, sha),
        message: branchRun.vcs_connection_commit?.message,
        author: branchRun.vcs_connection_commit?.author_name,
        createdAt:
          branchRun.vcs_connection_commit?.created_at ||
          branchRun.created_at ||
          run.created_at,
      }
    }
  }

  return found
}
