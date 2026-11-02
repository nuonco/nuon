import type {
  TAppBranchPreviewConfig,
  TAppBranchRunPreviewMode,
} from '@/types'

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
  author_login?: string
  html_url: string
  repo_owner: string
  repo_name: string
  head_ref: string
  base_ref: string
  head_sha: string

  state: TPullRequestState
  is_draft: boolean

  opened_at?: string
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

  added?: number
  changed?: number
  removed?: number
  components_built?: number
}

export type TPlaygroundInstallCandidate = {
  id: string
  name: string
  labels: Record<string, string>
  cloud_platform: string
  region: string
}
