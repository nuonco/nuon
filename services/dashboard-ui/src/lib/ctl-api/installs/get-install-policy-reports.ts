import { api } from '@/lib/api'
import type { TPaginationParams, TPolicyReport } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallPolicyReports = ({
  installId,
  orgId,
  ownerType,
  limit,
  offset,
}: {
  installId: string
  orgId: string
  ownerType?: string
} & TPaginationParams) =>
  api<TPolicyReport[]>({
    path: `policy-reports${buildQueryParams({ install_id: installId, owner_type: ownerType, limit, offset })}`,
    orgId,
  })
