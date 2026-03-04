import { api } from '@/lib/api'
import type { TInstallConfig } from '@/types'

export type TCreateInstallConfigBody = {
  approval_option: 'approve-all' | 'prompt'
}

export async function createInstallConfig({
  body,
  installId,
  orgId,
}: {
  body: TCreateInstallConfigBody
  installId: string
  orgId: string
}) {
  return api<TInstallConfig>({
    body,
    method: 'POST',
    orgId,
    path: `installs/${installId}/configs`,
  })
}
