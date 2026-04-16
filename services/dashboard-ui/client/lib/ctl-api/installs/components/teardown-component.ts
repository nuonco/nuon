import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export type TTeardownComponentBody = {
  error_behavior?: 'continue' | 'abort'
  plan_only?: boolean
}

export const teardownComponent = ({
  componentId,
  installId,
  orgId,
  body,
}: {
  componentId: string
  installId: string
  orgId: string
  body: TTeardownComponentBody
}) =>
  api<TWorkflowResponse>({
    withHeaders: true,
    path: `installs/${installId}/components/${componentId}/teardown`,
    method: 'POST',
    orgId,
    body,
  })
