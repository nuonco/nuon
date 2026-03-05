import { api } from '@/lib/api'

export async function shutdownRunnerInstance({
  orgId,
  runnerId,
}: {
  orgId: string
  runnerId: string
}) {
  return api<boolean>({
    body: {},
    method: 'POST',
    orgId,
    path: `runners/${runnerId}/mng/shutdown-vm`,
  })
}
