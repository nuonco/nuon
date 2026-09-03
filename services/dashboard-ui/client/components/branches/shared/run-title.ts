import type { TInstallWorkflow } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

export const WORKFLOW_TYPE_LABELS: Record<string, string> = {
  app_branches_manual_update: 'Manual app config update',
  app_branches_config_repo_update: 'Config update',
  app_branches_component_repo_update: 'Component update',
  app_branch_config_update: 'Config update',
}

export const getRunTitle = (run?: TInstallWorkflow): string => {
  const branchRun = run?.app_branch_runs?.at(0)
  if (branchRun?.pr_number != null) {
    return `PR #${branchRun.pr_number}`
  }

  const commitMessage = branchRun?.vcs_connection_commit?.message
    ?.split('\n')[0]
    ?.trim()
  const workflowName = run?.name === 'Manual run' ? 'Run' : run?.name
  const typeLabel = run?.type ? WORKFLOW_TYPE_LABELS[run.type] : undefined
  return toSentenceCase(
    commitMessage || workflowName || typeLabel || 'Workflow run'
  )
}
