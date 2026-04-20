import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export type TReprovisionSandboxBody = {
  plan_only: boolean
  skip_components?: boolean
}

export async function reprovisionSandbox({
  body,
  installId,
  orgId,
}: {
  body: TReprovisionSandboxBody
  installId: string
  orgId: string
}) {
  return api<TWorkflowResponse>({
    withHeaders: true,
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/reprovision-sandbox`,
  })
}
