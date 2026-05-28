import { api } from '@/lib/api'

export type TRunbookRun = {
  id: string
  runbook_id?: string
  install_id?: string
  org_id?: string
  created_at?: string
  updated_at?: string
  status?: string
  status_description?: string
  status_v2?: {
    status?: string
    status_human_description?: string
  }
  created_by_id?: string
  created_by?: {
    email?: string
    name?: string
  }
}

export const getInstallRunbookRun = ({
  installId,
  runId,
  orgId,
}: {
  installId: string
  runId: string
  orgId: string
}) =>
  api<TRunbookRun>({
    path: `installs/${installId}/runbook-runs/${runId}`,
    orgId,
  })
