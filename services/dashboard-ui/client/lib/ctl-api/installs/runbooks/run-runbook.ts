import { api } from '@/lib/api'
import type { TInstallRunbookRun } from './get-install-runbooks'

export async function runRunbook({
  installId,
  runbookId,
  orgId,
}: {
  installId: string
  runbookId: string
  orgId: string
}) {
  return api<TInstallRunbookRun>({
    method: 'POST',
    orgId,
    path: `installs/${installId}/runbooks/${runbookId}/runs`,
  })
}
