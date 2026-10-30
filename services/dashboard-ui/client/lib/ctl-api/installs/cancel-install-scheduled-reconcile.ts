import { api } from '@/lib/api'
import type { TInstallScheduledReconcile } from '@/types'

export async function cancelInstallScheduledReconcile({
  installId,
  orgId,
  scheduledReconcileId,
}: {
  installId: string
  orgId: string
  scheduledReconcileId: string
}) {
  return api<TInstallScheduledReconcile>({
    method: 'DELETE',
    orgId,
    path: `installs/${installId}/scheduled-reconciles/${scheduledReconcileId}`,
  })
}
