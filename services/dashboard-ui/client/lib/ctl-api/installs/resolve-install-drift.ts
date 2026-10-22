import { api } from '@/lib/api'
import type { TResolveInstallDriftBody, TWorkflowResponse } from '@/types'

export const resolveInstallDrift = ({
  body,
  installId,
  orgId,
}: {
  body: TResolveInstallDriftBody
  installId: string
  orgId: string
}) =>
  api<TWorkflowResponse>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/resolve-drift`,
  })
