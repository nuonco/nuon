import { api } from '@/lib/api'

export async function shutdownRunner({
  force = false,
  orgId,
  runnerId,
}: {
  force?: boolean
  orgId: string
  runnerId: string
}) {
  const path = force
    ? `runners/${runnerId}/force-shutdown`
    : `runners/${runnerId}/graceful-shutdown`

  return api<boolean>({
    body: {},
    method: 'POST',
    orgId,
    path,
  })
}
