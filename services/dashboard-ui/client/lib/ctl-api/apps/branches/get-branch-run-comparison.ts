import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'

export type TBranchRunComparisonConfigDiffEntry = {
  op: string
  name: string
  description?: string
  source_changed?: boolean
  file?: string
}

export type TBranchRunComparisonConfigDiffSection = {
  name: string
  additions: number
  removals: number
  changed: number
  entries: TBranchRunComparisonConfigDiffEntry[]
}

export type TBranchRunComparisonConfigDiff = {
  config_file?: string
  additions: number
  removals: number
  changed: number
  sections: TBranchRunComparisonConfigDiffSection[]
}

export type TBranchRunComparisonRunSummary = {
  id: string
  workflow_id?: string
  status?: string
  created_at?: string
  pr_number?: number
  base_branch?: string
  event_type?: string
  app_config_id?: string
  vcs_connection_commit?: {
    sha?: string
    message?: string
    author_name?: string
    author_avatar_url?: string
  }
}

export type TSourceArchiveFileDiff = {
  path: string
  op: string
  before_sha256?: string
  after_sha256?: string
  before_size?: number
  after_size?: number
  patch?: string
  patch_truncated?: boolean
}

export type TSourceArchiveDiff = {
  total_files: number
  unchanged: number
  files: TSourceArchiveFileDiff[]
}

export type TBranchRunComparison = {
  id: string
  head_run_id: string
  base_run_id?: string
  base_sha?: string
  head_sha?: string
  head_run?: TBranchRunComparisonRunSummary
  base_run?: TBranchRunComparisonRunSummary
  git_diff_content?: unknown
  full_diff_content?: unknown
  config_diff_content?: TBranchRunComparisonConfigDiff
  source_diff_content?: TSourceArchiveDiff
}

export const getBranchRunComparison = ({
  appId,
  branchId,
  runId,
  orgId,
  includeDiff,
}: {
  appId: string
  branchId: string
  runId: string
  orgId: string
  includeDiff?: Array<'git' | 'full' | 'config' | 'source'>
}) =>
  api<TBranchRunComparison>({
    path: `apps/${appId}/branches/${branchId}/runs/${runId}/comparison${buildQueryParams({
      include_diff: includeDiff?.length ? includeDiff.join(',') : undefined,
    })}`,
    orgId,
  })
