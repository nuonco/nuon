import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export type TRunAdhocActionBody = {
  command?: string
  env_vars?: Record<string, string>
  inline_contents?: string
  name?: string
  role?: string
  timeout?: number
}

export async function runAdhocAction({
  body,
  installId,
  orgId,
}: {
  body: TRunAdhocActionBody
  installId: string
  orgId: string
}) {
  return api<TWorkflowResponse>({
    withHeaders: true,
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/actions/adhoc-run`,
  })
}
