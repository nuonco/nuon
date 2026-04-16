import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export type TDeprovisionInstallBody = {
  error_behavior?: 'continue' | 'abort'
  plan_only: boolean
}

export async function deprovisionInstall({
  body,
  installId,
  orgId,
}: {
  body: TDeprovisionInstallBody
  installId: string
  orgId: string
}) {
  return api<TWorkflowResponse>({
    withHeaders: true,
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/deprovision`,
  })
}
