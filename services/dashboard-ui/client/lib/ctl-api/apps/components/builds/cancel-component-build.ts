import { api } from '@/lib/api'
import type { TBuild } from '@/types'

export async function cancelComponentBuild({
  orgId,
  appId,
  componentId,
  buildId,
}: {
  orgId: string
  appId: string
  componentId: string
  buildId: string
}) {
  return api<TBuild>({
    method: 'POST',
    orgId,
    path: `apps/${appId}/components/${componentId}/builds/${buildId}/cancel`,
  })
}
