import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export type TToggleComponentBody = {
  enabled: boolean
  plan_only?: boolean
  role?: string
  tear_down_dependents?: boolean
}

export const toggleComponent = ({
  componentId,
  installId,
  orgId,
  body,
}: {
  componentId: string
  installId: string
  orgId: string
  body: TToggleComponentBody
}) =>
  api<TWorkflowResponse>({
    withHeaders: true,
    path: `installs/${installId}/components/${componentId}/toggle`,
    method: 'POST',
    orgId,
    body,
  })
