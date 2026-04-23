import { api } from '@/lib/api'
import type { TInstall } from '@/types'

export async function addInstallLabels({
  body,
  installId,
  orgId,
}: {
  body: { labels: Record<string, string> }
  installId: string
  orgId: string
}) {
  return api<TInstall>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/labels`,
  })
}
