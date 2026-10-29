import { api } from '@/lib/api'
import type { TInstallAppConfigVersion } from '@/types'

export type TReconcileInstallAppVersionBody = {
  plan_only: boolean
  request_id?: string
}

export async function reconcileInstallAppVersion({
  body,
  installId,
  orgId,
}: {
  body: TReconcileInstallAppVersionBody
  installId: string
  orgId: string
}) {
  return api<TInstallAppConfigVersion>({
    withHeaders: true,
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/reconcile-app-version`,
  })
}
