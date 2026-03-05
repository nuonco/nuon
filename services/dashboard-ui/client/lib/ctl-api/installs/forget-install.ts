import { api } from '@/lib/api'

export async function forgetInstall({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) {
  return api<boolean>({
    method: 'POST',
    orgId,
    path: `installs/${installId}/forget`,
  })
}
