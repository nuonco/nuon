import type { TInstallDeploymentRecord, TWorkflow } from '@/types'
import { vcsRepo } from '@/utils/vcs-urls'

export type TDeploymentRunBranch = {
  id: string
  name?: string
  runId?: string
  status?: string
  sha?: string
  message?: string
  author?: string
  avatarUrl?: string
  createdAt?: string
  repo?: string
}

const runBranchId = (run: NonNullable<TWorkflow['app_branch_runs']>[number]) =>
  run.app_branch?.id || (run as { app_branch_id?: string }).app_branch_id

// The deployment record falls back to the install's tracking branch when a
// workflow has no branch run of its own. That ref has no run id.
export const deploymentRunBranch = (
  deployment?: Pick<TInstallDeploymentRecord, 'app_branch'> | null,
  workflow?: Pick<TWorkflow, 'app_branch_runs'> | null
): TDeploymentRunBranch | undefined => {
  const run = workflow?.app_branch_runs?.[0]
  const id = run ? runBranchId(run) : undefined
  if (run && id) {
    const commit = run.vcs_connection_commit
    return {
      id,
      name: run.app_branch?.name || undefined,
      runId: run.id,
      status: run.status,
      sha: commit?.sha || run.head_sha || run.metadata?.head_sha || undefined,
      message: commit?.message?.split('\n')[0] || undefined,
      author: commit?.author_name || undefined,
      avatarUrl: commit?.author_avatar_url || undefined,
      createdAt: run.created_at,
      repo: vcsRepo(run.app_branch_config),
    }
  }

  const ref = deployment?.app_branch
  if (ref?.id && ref.run_id) {
    return { id: ref.id, name: ref.name, runId: ref.run_id, sha: ref.sha }
  }

  return undefined
}
