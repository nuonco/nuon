import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export const recoverHelmRelease = ({
  componentId,
  installId,
  orgId,
}: {
  componentId: string
  installId: string
  orgId: string
}) =>
  api<TWorkflowResponse>({
    withHeaders: true,
    path: `installs/${installId}/components/${componentId}/recover-helm-release`,
    method: 'POST',
    orgId,
    body: {},
  })
