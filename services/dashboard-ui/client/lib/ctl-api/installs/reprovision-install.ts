import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export type TReprovisionInstallBody = {
  plan_only: boolean
  role?: string
}

export async function reprovisionInstall({
  body,
  installId,
  orgId,
}: {
  body: TReprovisionInstallBody
  installId: string
  orgId: string
}) {
  return api<TWorkflowResponse>({
    withHeaders: true,
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/reprovision`,
  })
}
