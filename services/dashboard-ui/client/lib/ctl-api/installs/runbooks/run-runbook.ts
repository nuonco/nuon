import { api } from '@/lib/api'
import type { TInstallRunbookRun } from './get-install-runbooks'

export type TRunRunbookBody = {
  inputs?: Record<string, string>
}

export async function runRunbook({
  installId,
  runbookId,
  orgId,
  body,
}: {
  installId: string
  runbookId: string
  orgId: string
  body?: TRunRunbookBody
}) {
  return api<TInstallRunbookRun>({
    method: 'POST',
    orgId,
    path: `installs/${installId}/runbooks/${runbookId}/runs`,
    ...(body ? { body } : {}),
  })
}
