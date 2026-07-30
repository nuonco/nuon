import { api } from '@/lib/api'

export async function resetInstallHealthBaseline({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) {
  return api<{ baseline_at: string }>({
    method: 'POST',
    orgId,
    path: `installs/${installId}/health/baseline`,
  })
}
