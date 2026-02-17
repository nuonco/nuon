import { api } from '@/lib/api'
import type { TPaginationParams, TPolicyReport } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export type TPolicyReportStatus = 'error' | 'warning' | 'success'
export type TPolicyReportOwnerType =
  | 'install_deploys'
  | 'install_sandbox_runs'
  | 'component_builds'

export const getInstallPolicyReports = ({
  installId,
  orgId,
  ownerType,
  status,
  limit,
  offset,
}: {
  installId: string
  orgId: string
  ownerType?: TPolicyReportOwnerType
  status?: TPolicyReportStatus
} & TPaginationParams) =>
  api<TPolicyReport[]>({
    path: `policy-reports${buildQueryParams({ install_id: installId, owner_type: ownerType, status, limit, offset })}`,
    orgId,
  })
