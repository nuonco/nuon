import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'

export type TBranchRunComparisonConfigDiffEntry = {
  op: string
  name: string
  description?: string
  source_changed?: boolean
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
  vcs_connection_commit?: {
    sha?: string
    message?: string
    author_name?: string
    author_avatar_url?: string
  }
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
  includeDiff?: Array<'git' | 'full' | 'config'>
}) =>
  api<TBranchRunComparison>({
    path: `apps/${appId}/branches/${branchId}/runs/${runId}/comparison${buildQueryParams(
      {
        include_diff: includeDiff?.length ? includeDiff.join(',') : undefined,
      }
    )}`,
    orgId,
  })
