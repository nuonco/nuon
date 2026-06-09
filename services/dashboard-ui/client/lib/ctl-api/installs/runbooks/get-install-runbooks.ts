import { api } from '@/lib/api'
import type { TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks'

export type TInstallRunbook = {
  id: string
  install_id?: string
  runbook_id?: string
  org_id?: string
  created_at?: string
  updated_at?: string
  status?: string
  runbook: TRunbook
  runs?: TInstallRunbookRun[]
}

export type TInstallRunbookRun = {
  id: string
  install_id?: string
  install_runbook_id?: string
  runbook_config_id?: string
  org_id?: string
  created_at?: string
  updated_at?: string
  status?: string
  status_description?: string
  triggered_by_id?: string
  created_by_id?: string
  created_by?: {
    email?: string
    name?: string
  }
  install_workflow_id?: string | null
  install_workflow?: {
    id: string
    status?: {
      status?: string
      status_human_description?: string
    }
  } | null
}

export async function getInstallRunbooks({
  installId,
  orgId,
  limit,
  offset,
  q,
}: {
  installId: string
  orgId: string
  q?: string
} & TPaginationParams) {
  return api<TInstallRunbook[]>({
    orgId,
    path: `installs/${installId}/runbooks${buildQueryParams({ limit, offset, q })}`,
    paginated: true,
  })
}
