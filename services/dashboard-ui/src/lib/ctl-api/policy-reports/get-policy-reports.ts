import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'
import type { TPolicyReport } from '@/types/ctl-api.types'

export type TPolicyReportOwnerType =
  | 'install_deploys'
  | 'install_sandbox_runs'
  | 'component_builds'

export type TPolicyReportStatus =
  | 'pending'
  | 'in_progress'
  | 'success'
  | 'error'
  | 'cancelled'
  | 'skipped'
  | 'paused'
  | 'not_attempted'
  | 'warning'

export const getPolicyReports = ({
  orgId,
  ownerType,
  ownerId,
  appId,
  installId,
  status,
  limit,
  offset,
}: {
  orgId: string
  ownerType?: TPolicyReportOwnerType
  ownerId?: string
  appId?: string
  installId?: string
  status?: TPolicyReportStatus
  limit?: number
  offset?: number
}) =>
  api<TPolicyReport[]>({
    path: `policy-reports${buildQueryParams({
      owner_type: ownerType,
      owner_id: ownerId,
      app_id: appId,
      install_id: installId,
      status,
      limit,
      offset,
    })}`,
    orgId,
  })
