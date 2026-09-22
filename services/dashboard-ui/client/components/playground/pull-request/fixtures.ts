import type {
  TAppBranchPreviewConfig,
  TAppBranchRunPreviewMode,
} from '@/types'

export const PR_ORG_ID = 'org-pr-scope-demo'
export const PR_APP_ID = 'app-acme-platform'
export const PR_BRANCH_ID = 'brnch-main'

/**
 * Mirrors the proposed AppBranchPullRequest model in
 * specs/app-branch-pull-requests.md. The PR is scoped to
 * (app_branch_id, number) and owns the preview override that every run on the
 * PR resolves against.
 */
export type TPullRequestState = 'open' | 'closed' | 'merged'

export type TPullRequestConfigStatus =
  | 'ready-from-defaults'
  | 'needs-input'
  | 'configured'

export type TPlaygroundPullRequest = {
  id: string
  org_id: string
  app_branch_id: string
  app_id: string
  number: number

  title: string
  author_login: string
  html_url: string
  repo_owner: string
  repo_name: string
  head_ref: string
  base_ref: string
  head_sha: string

  state: TPullRequestState
  is_draft: boolean

  opened_at: string
  closed_at?: string
  merged_at?: string

  config_status: TPullRequestConfigStatus
  branch_preview_config: TAppBranchPreviewConfig
  preview_override?: TAppBranchPreviewConfig
  resolved_preview_config: TAppBranchPreviewConfig
}

export type TPlaygroundPullRequestRun = {
  id: string
  run_number: number
  status: 'pending' | 'running' | 'success' | 'failed' | 'cancelled'
  trigger: 'pull_request' | 'push' | 'manual'

  mode: TAppBranchRunPreviewMode
  install_id?: string
  install_name?: string

  head_sha: string
  commit_message: string
  commit_author: string

  started_at: string
  completed_at?: string
  error_message?: string

  /** Counts from the run's comparison, rendered as the impact summary. */
  added: number
  changed: number
  removed: number
  components_built: number
}

export type TPlaygroundInstallCandidate = {
  id: string
  name: string
  labels: Record<string, string>
  cloud_platform: string
  region: string
}

export const installCandidates: TPlaygroundInstallCandidate[] = [
  {
    id: 'insta1b2c3d4e5f6g7h8i9j0k1',
    name: 'acme-preview',
    labels: { env: 'preview', tier: 'internal' },
    cloud_platform: 'aws',
    region: 'us-east-1',
  },
  {
    id: 'instb2c3d4e5f6g7h8i9j0k1l2',
    name: 'acme-staging',
    labels: { env: 'staging', tier: 'internal' },
    cloud_platform: 'aws',
    region: 'us-west-2',
  },
  {
    id: 'instc3d4e5f6g7h8i9j0k1l2m3',
    name: 'acme-sandbox',
    labels: { env: 'sandbox', tier: 'internal' },
    cloud_platform: 'gcp',
    region: 'us-central1',
  },
  {
    id: 'instd4e5f6g7h8i9j0k1l2m3n4',
    name: 'acme-production',
    labels: { env: 'prod', tier: 'enterprise' },
    cloud_platform: 'aws',
    region: 'us-east-1',
  },
]

/**
 * Branch defaults that fail AppBranchPreviewConfig.Validate() — a non
 * build-only mode with no install target. Drives config_status 'needs-input'.
 */
export const incompleteBranchPreviewConfig: TAppBranchPreviewConfig = {
  mode: 'plan-only',
  set_statuses: true,
  comment: true,
  ignore_drafts: true,
}

/** Branch defaults that already name an install, so runs just work. */
export const completeBranchPreviewConfig: TAppBranchPreviewConfig = {
  ...incompleteBranchPreviewConfig,
  install_id: 'instb2c3d4e5f6g7h8i9j0k1l2',
  install_name: 'acme-staging',
}

const basePullRequest = {
  id: 'abprq1b2c3d4e5f6g7h8i9j0k1',
  org_id: PR_ORG_ID,
  app_branch_id: PR_BRANCH_ID,
  app_id: PR_APP_ID,
  number: 482,
  title: 'Bump the api component to node 22 and widen the worker pool',
  author_login: 'dana-eng',
  html_url: 'https://github.com/acme/acme-platform/pull/482',
  repo_owner: 'acme',
  repo_name: 'acme-platform',
  head_ref: 'dana/node-22-upgrade',
  base_ref: 'main',
  head_sha: '9f2c1ab7e4d05c3188a6b0f7e2d9c4a1b3e5f701',
  state: 'open' as TPullRequestState,
  is_draft: false,
  opened_at: '2026-09-21T14:02:00Z',
}

/** First visit: branch defaults are incomplete, so the PR needs an install. */
export const needsInputPullRequest: TPlaygroundPullRequest = {
  ...basePullRequest,
  config_status: 'needs-input',
  branch_preview_config: incompleteBranchPreviewConfig,
  resolved_preview_config: incompleteBranchPreviewConfig,
}

/** First visit: branch defaults already resolve, so the run just ran. */
export const readyFromDefaultsPullRequest: TPlaygroundPullRequest = {
  ...basePullRequest,
  config_status: 'ready-from-defaults',
  branch_preview_config: completeBranchPreviewConfig,
  resolved_preview_config: completeBranchPreviewConfig,
}

/** Return visit: someone saved a PR-scoped override. */
export const configuredPullRequest: TPlaygroundPullRequest = {
  ...basePullRequest,
  config_status: 'configured',
  branch_preview_config: incompleteBranchPreviewConfig,
  preview_override: {
    mode: 'apply',
    install_id: 'insta1b2c3d4e5f6g7h8i9j0k1',
    install_name: 'acme-preview',
  },
  resolved_preview_config: {
    mode: 'apply',
    install_id: 'insta1b2c3d4e5f6g7h8i9j0k1',
    install_name: 'acme-preview',
    set_statuses: true,
    comment: true,
    ignore_drafts: true,
  },
}

export const mergedPullRequest: TPlaygroundPullRequest = {
  ...configuredPullRequest,
  state: 'merged',
  merged_at: '2026-09-22T11:40:00Z',
  closed_at: '2026-09-22T11:40:00Z',
}

export const draftPullRequest: TPlaygroundPullRequest = {
  ...basePullRequest,
  is_draft: true,
  config_status: 'ready-from-defaults',
  branch_preview_config: { ...completeBranchPreviewConfig, mode: 'build-only' },
  resolved_preview_config: {
    ...completeBranchPreviewConfig,
    mode: 'build-only',
  },
}

/**
 * Five preview runs across three commits — the history requirement. The first
 * two ran build-only because the PR had no install yet; the rest ran at the
 * mode chosen on the PR, which is why each run keeps its own resolved mode.
 */
export const previewRuns: TPlaygroundPullRequestRun[] = [
  {
    id: 'abrun5f6g7h8i9j0k1l2m3n4o5',
    run_number: 5,
    status: 'success',
    trigger: 'pull_request',
    mode: 'apply',
    install_id: 'insta1b2c3d4e5f6g7h8i9j0k1',
    install_name: 'acme-preview',
    head_sha: '9f2c1ab7e4d05c3188a6b0f7e2d9c4a1b3e5f701',
    commit_message: 'Widen the worker pool to 8 and drop the legacy shim',
    commit_author: 'dana-eng',
    started_at: '2026-09-22T10:31:00Z',
    completed_at: '2026-09-22T10:37:20Z',
    added: 2,
    changed: 4,
    removed: 1,
    components_built: 3,
  },
  {
    id: 'abrun4e5f6g7h8i9j0k1l2m3n4',
    run_number: 4,
    status: 'failed',
    trigger: 'pull_request',
    mode: 'apply',
    install_id: 'insta1b2c3d4e5f6g7h8i9j0k1',
    install_name: 'acme-preview',
    head_sha: '9f2c1ab7e4d05c3188a6b0f7e2d9c4a1b3e5f701',
    commit_message: 'Widen the worker pool to 8 and drop the legacy shim',
    commit_author: 'dana-eng',
    started_at: '2026-09-22T09:58:00Z',
    completed_at: '2026-09-22T10:04:45Z',
    error_message:
      'helm upgrade failed: timed out waiting for the condition on deployment/api',
    added: 0,
    changed: 0,
    removed: 0,
    components_built: 3,
  },
  {
    id: 'abrun3d4e5f6g7h8i9j0k1l2m3',
    run_number: 3,
    status: 'success',
    trigger: 'pull_request',
    mode: 'plan-only',
    install_id: 'insta1b2c3d4e5f6g7h8i9j0k1',
    install_name: 'acme-preview',
    head_sha: '4c8de91b2a7f60d3e5c1908baf27d46e0b39a5c2',
    commit_message: 'Pin the node base image to 22.7-alpine',
    commit_author: 'dana-eng',
    started_at: '2026-09-22T08:12:00Z',
    completed_at: '2026-09-22T08:16:10Z',
    added: 2,
    changed: 4,
    removed: 1,
    components_built: 3,
  },
  {
    id: 'abrun2c3d4e5f6g7h8i9j0k1l2',
    run_number: 2,
    status: 'success',
    trigger: 'pull_request',
    mode: 'build-only',
    head_sha: '4c8de91b2a7f60d3e5c1908baf27d46e0b39a5c2',
    commit_message: 'Pin the node base image to 22.7-alpine',
    commit_author: 'dana-eng',
    started_at: '2026-09-21T16:45:00Z',
    completed_at: '2026-09-21T16:49:30Z',
    added: 0,
    changed: 0,
    removed: 0,
    components_built: 3,
  },
  {
    id: 'abrun1b2c3d4e5f6g7h8i9j0k1',
    run_number: 1,
    status: 'success',
    trigger: 'pull_request',
    mode: 'build-only',
    head_sha: 'e71ba3c0d8492f5617ae3c2b90d4f8a15c6e0d33',
    commit_message: 'Upgrade api component to node 22',
    commit_author: 'dana-eng',
    started_at: '2026-09-21T14:03:00Z',
    completed_at: '2026-09-21T14:08:05Z',
    added: 0,
    changed: 0,
    removed: 0,
    components_built: 3,
  },
]

/** The single run that fires from branch defaults before anyone configures. */
export const firstRunOnly: TPlaygroundPullRequestRun[] = [previewRuns[4]]

export const runningRun: TPlaygroundPullRequestRun = {
  ...previewRuns[0],
  id: 'abrun6g7h8i9j0k1l2m3n4o5p6',
  run_number: 6,
  status: 'running',
  started_at: '2026-09-22T12:01:00Z',
  completed_at: undefined,
  added: 0,
  changed: 0,
  removed: 0,
  components_built: 0,
}

export const installById = (id?: string) =>
  installCandidates.find((install) => install.id === id)

export const pullRequestPath = (pr: TPlaygroundPullRequest) =>
  `/${pr.org_id}/apps/${pr.app_id}/branches/${pr.app_branch_id}/pull-requests/${pr.number}`
