import type {
  TPlaygroundPullRequest,
  TPlaygroundPullRequestRun,
  TPullRequestConfigStatus,
} from './fixtures'
import type {
  TAppBranchPreviewConfig,
  TAppBranchRun,
  TInstallWorkflow,
  TPreviewSourcePR,
} from '@/types'

const RUNNING = new Set([
  'in-progress',
  'checking-plan',
  'retrying',
  'planning',
  'applying',
  'generating',
  'provisioning',
  'building',
])
const SUCCESS = new Set([
  'success',
  'active',
  'approved',
  'noop',
  'no-drift',
  'finished',
  'succeeded',
])
const FAILED = new Set(['error', 'failed', 'failed-pending-retry'])
const CANCELLED = new Set([
  'cancelled',
  'discarded',
  'user-skipped',
  'auto-skipped',
  'disabled',
])

export const previewRunStatus = (
  status?: string
): TPlaygroundPullRequestRun['status'] => {
  if (!status || status === 'pending' || status === 'queued') return 'pending'
  if (RUNNING.has(status)) return 'running'
  if (SUCCESS.has(status)) return 'success'
  if (FAILED.has(status)) return 'failed'
  if (CANCELLED.has(status)) return 'cancelled'
  return 'pending'
}

const hasInstallTarget = (config?: TAppBranchPreviewConfig) =>
  !!config?.install_id ||
  !!config?.install_name ||
  !!Object.keys(config?.label_selector?.match_labels ?? {}).length

export const previewConfigStatus = (
  branchConfig: TAppBranchPreviewConfig | undefined,
  latest?: TAppBranchRun
): TPullRequestConfigStatus => {
  const override = latest?.preview?.override_preview_config
  if (override?.mode || override?.install_id) return 'configured'
  const resolved =
    latest?.preview?.resolved_preview_config ?? branchConfig ?? {}
  const mode = resolved.mode ?? 'plan-only'
  if (mode !== 'build-only' && mode !== 'none' && !hasInstallTarget(resolved)) {
    return 'needs-input'
  }
  return 'ready-from-defaults'
}

const triggerOf = (
  trigger?: string
): TPlaygroundPullRequestRun['trigger'] => {
  if (trigger === 'pull_request' || trigger === 'push' || trigger === 'manual') {
    return trigger
  }
  return 'manual'
}

const repoParts = (repo?: string) => {
  const trimmed = repo?.trim() ?? ''
  const parts = trimmed.split('/').filter(Boolean)
  if (parts.length >= 2) {
    return {
      repo_owner: parts[parts.length - 2],
      repo_name: parts[parts.length - 1],
    }
  }
  return { repo_owner: '', repo_name: trimmed }
}

export const toPlaygroundRun = (
  workflow: TInstallWorkflow,
  runNumber: number
): TPlaygroundPullRequestRun | undefined => {
  const branchRun = workflow.app_branch_runs?.[0]
  if (!branchRun) return undefined
  const commit = branchRun.vcs_connection_commit
  const preview = branchRun.preview
  const resolved = preview?.resolved_preview_config
  const mode = preview?.mode ?? resolved?.mode ?? 'plan-only'
  const status = previewRunStatus(workflow.status?.status || branchRun.status)
  const started =
    branchRun.started_at || branchRun.created_at || workflow.created_at
  if (!started) return undefined

  return {
    id: branchRun.id || workflow.id || `run-${runNumber}`,
    run_number: runNumber,
    status,
    trigger: triggerOf(branchRun.metadata?.trigger ?? branchRun.event_type),
    mode,
    install_id: preview?.install_id ?? resolved?.install_id,
    install_name: preview?.install_name ?? resolved?.install_name,
    head_sha: commit?.sha || branchRun.head_sha || branchRun.metadata?.head_sha || '',
    commit_message: commit?.message || workflow.name || 'Preview run',
    commit_author: commit?.author_name || '',
    started_at: started,
    completed_at: workflow.finished_at || branchRun.completed_at,
    error_message:
      status === 'failed'
        ? branchRun.error_message || workflow.status?.status_human_description
        : branchRun.error_message,
    components_built: undefined,
  }
}

export const pullRequestRuns = (
  workflows: TInstallWorkflow[],
  prNumber: number
) => {
  const matched = workflows.filter((workflow) => {
    const branchRun = workflow.app_branch_runs?.[0]
    const number = branchRun?.pr_number ?? branchRun?.metadata?.pr_number
    return number === prNumber
  })
  const chronological = [...matched].sort(
    (a, b) =>
      Date.parse(a.created_at || a.app_branch_runs?.[0]?.created_at || '') -
      Date.parse(b.created_at || b.app_branch_runs?.[0]?.created_at || '')
  )
  return chronological
    .map((workflow, index) => toPlaygroundRun(workflow, index + 1))
    .filter((run): run is TPlaygroundPullRequestRun => !!run)
    .reverse()
}

export const toPlaygroundPullRequest = ({
  source,
  orgId,
  appId,
  branchId,
  repo,
  baseRef,
  branchConfig,
  runs,
  latest,
}: {
  source: TPreviewSourcePR
  orgId: string
  appId: string
  branchId: string
  repo?: string
  baseRef?: string
  branchConfig?: TAppBranchPreviewConfig
  runs: TPlaygroundPullRequestRun[]
  latest?: TAppBranchRun
}): TPlaygroundPullRequest => {
  const names = repoParts(repo)
  const branchPreview =
    latest?.preview?.branch_preview_config ?? branchConfig ?? {}
  const resolved =
    latest?.preview?.resolved_preview_config ?? branchPreview
  const override = latest?.preview?.override_preview_config
  const headSha = source.head_sha || runs[0]?.head_sha || ''

  return {
    id: `pr-${source.pr_number}`,
    org_id: orgId,
    app_branch_id: branchId,
    app_id: appId,
    number: source.pr_number,
    title: source.title || `Pull request #${source.pr_number}`,
    author_login: runs[0]?.commit_author || undefined,
    html_url: source.url || '',
    ...names,
    head_ref: source.head_ref || '',
    base_ref: baseRef || '',
    head_sha: headSha,
    state: 'open',
    is_draft: !!latest?.metadata?.is_draft || !!latest?.preview?.is_draft_mode,
    opened_at: [...runs].at(-1)?.started_at,
    config_status: previewConfigStatus(branchConfig, latest),
    branch_preview_config: branchPreview,
    preview_override: override
      ? {
          mode: override.mode,
          install_id: override.install_id,
          install_name: resolved.install_name,
        }
      : undefined,
    resolved_preview_config: resolved,
  }
}
