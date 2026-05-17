import { api } from '@/lib/api'
import type { TWorkflowResponse } from '@/types'

export async function runRunbook({
  installId,
  runbookId,
  orgId,
}: {
  installId: string
  runbookId: string
  orgId: string
}) {
  return api<TWorkflowResponse>({
    method: 'POST',
    orgId,
    path: `installs/${installId}/runbooks/${runbookId}/runs`,
  })
}
