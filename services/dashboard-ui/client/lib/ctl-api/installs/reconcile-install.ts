import { api } from '@/lib/api'
import type { TReconcileInstallResponse } from '@/types'

export type TReconcileInstallBody = {
  expected_app_config_id: string
  plan_only?: boolean
}

export const reconcileInstall = ({
  body,
  installId,
  orgId,
}: {
  body: TReconcileInstallBody
  installId: string
  orgId: string
}) =>
  api<TReconcileInstallResponse>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/reconcile`,
  })
