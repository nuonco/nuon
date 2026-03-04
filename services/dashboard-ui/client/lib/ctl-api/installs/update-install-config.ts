import { api } from '@/lib/api'

export type TUpdateInstallConfigBody = {
  approval_option: 'approve-all' | 'prompt'
}

export async function updateInstallConfig({
  body,
  installConfigId,
  installId,
  orgId,
}: {
  body: TUpdateInstallConfigBody
  installConfigId: string
  installId: string
  orgId: string
}) {
  return api<string>({
    body,
    method: 'PATCH',
    orgId,
    path: `installs/${installId}/configs/${installConfigId}`,
  })
}
