import { api } from '@/lib/api'

export async function forgetComponent({
  componentId,
  installId,
  orgId,
}: {
  componentId: string
  installId: string
  orgId: string
}) {
  return api<boolean>({
    method: 'POST',
    orgId,
    path: `installs/${installId}/components/${componentId}/forget`,
  })
}
