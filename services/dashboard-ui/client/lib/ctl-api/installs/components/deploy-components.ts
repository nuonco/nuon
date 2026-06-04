import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export type TDeployComponentsBody = {
  plan_only?: boolean
  role?: string
}

export const deployComponents = ({
  installId,
  orgId,
  body,
}: {
  installId: string
  orgId: string
  body: TDeployComponentsBody
}) =>
  api<TWorkflowResponse>({
    withHeaders: true,
    path: `installs/${installId}/components/deploy-all`,
    method: 'POST',
    orgId,
    body,
  })
