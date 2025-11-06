import { api } from '@/lib/api'
import type { TInstallAuditLog } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallAuditLog = ({
  installId,
  orgId,
  start,
  end,
}: {
  installId: string
  orgId: string
  start: string
  end: string
}) =>
  api<TInstallAuditLog[]>({
    path: `installs/${installId}/audit_logs${buildQueryParams({ start, end })}`,
    orgId,
  })
