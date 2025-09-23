import { api } from '@/lib/api'
import type { TInstall } from '@/types'

export type TUpdateInstallBody = {
  install_config?: {
    approval_option: 'prompt' | 'approve-all'
  }
  metadata?: {
    managed_by: 'nuon/dashboard' // 'nuon/cli/config'
  }
}

export async function updateInstall({
  body,
  installId,
  orgId,
}: {
  body: TUpdateInstallBody
  installId: string
  orgId: string
}) {
  return api<TInstall>({
    body,
    method: 'PATCH',
    orgId,
    path: `installs/${installId}`,
  })
}
