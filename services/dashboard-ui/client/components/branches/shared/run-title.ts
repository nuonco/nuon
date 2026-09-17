import type { TAppBranchRun, TInstallWorkflow } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

export type TAppBranchRunTrigger =
  | 'manual'
  | 'push'
  | 'pull_request'
  | 'tag'
  | 'github_label'
  | 'onboarding'

export const WORKFLOW_TYPE_LABELS: Record<string, string> = {
  app_branches_manual_update: 'Manual app config update',
  app_branches_config_repo_update: 'Config update',
  app_branches_component_repo_update: 'Component update',
  app_branch_config_update: 'Config update',
}

export const getRunTrigger = (
  branchRun?: TAppBranchRun
): TAppBranchRunTrigger | undefined =>
  (branchRun?.metadata?.trigger ?? branchRun?.event_type) as
    | TAppBranchRunTrigger
    | undefined

export const getRunTitle = (run?: TInstallWorkflow): string => {
  const branchRun = run?.app_branch_runs?.at(0) as TAppBranchRun | undefined
  const metadata = branchRun?.metadata
  const prNumber = metadata?.pr_number ?? branchRun?.pr_number
  const trigger = getRunTrigger(branchRun)
  if (trigger === 'tag' && metadata?.tag) {
    return `Tag ${metadata.tag} Pushed`
  }
  if (trigger === 'github_label' && prNumber != null) {
    return `PR #${prNumber} Merged`
  }
  if (prNumber != null) {
    return `PR #${prNumber}`
  }
  if (trigger === 'github_label' && metadata?.github_label) {
    return `Label ${metadata.github_label}`
  }
  if (
    branchRun?.preview ||
    branchRun?.plan_only ||
    branchRun?.run_type === 'git-preview-run'
  ) {
    return 'Preview run'
  }
  if (trigger === 'manual') {
    return 'Manual run'
  }
  if (trigger === 'push') {
    return 'Commit pushed'
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
