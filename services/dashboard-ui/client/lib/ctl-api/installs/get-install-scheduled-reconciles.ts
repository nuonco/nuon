import { api } from '@/lib/api'
import type {
  TInstallScheduledReconcile,
  TInstallScheduledReconcileStatus,
} from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getInstallScheduledReconciles = ({
  installId,
  orgId,
  status,
  limit,
  offset,
}: {
  installId: string
  orgId: string
  status?: TInstallScheduledReconcileStatus
  limit?: number
  offset?: number
}) =>
  api<TInstallScheduledReconcile[]>({
    path: `installs/${installId}/scheduled-reconciles${buildQueryParams({
      status,
      limit,
      offset,
    })}`,
    orgId,
  })
