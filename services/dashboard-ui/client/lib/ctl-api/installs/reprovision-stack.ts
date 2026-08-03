import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export type TReprovisionStackBody = {
  plan_only: boolean
  skip_components?: boolean
  role?: string
}

export async function reprovisionStack({
  body,
  installId,
  orgId,
}: {
  body: TReprovisionStackBody
  installId: string
  orgId: string
}) {
  return api<TWorkflowResponse>({
    withHeaders: true,
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/reprovision-stack`,
  })
}
