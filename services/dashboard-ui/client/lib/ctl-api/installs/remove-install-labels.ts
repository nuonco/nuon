import { api } from '@/lib/api'
import type { TInstall } from '@/types'

export async function removeInstallLabels({
  body,
  installId,
  orgId,
}: {
  body: { keys: string[] }
  installId: string
  orgId: string
}) {
  return api<TInstall>({
    body,
    method: 'DELETE',
    orgId,
    path: `installs/${installId}/labels`,
  })
}
