import { api } from '@/lib/api'
import type { TInstallScheduledReconcile } from '@/types'

export type TCreateInstallScheduledReconcileBody = {
  scheduled_at: string
  plan_only?: boolean
}

export async function createInstallScheduledReconcile({
  body,
  installId,
  orgId,
}: {
  body: TCreateInstallScheduledReconcileBody
  installId: string
  orgId: string
}) {
  return api<TInstallScheduledReconcile>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/scheduled-reconciles`,
  })
}
