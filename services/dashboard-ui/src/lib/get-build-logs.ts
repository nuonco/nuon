import type { TComponentBuildLogs } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetBuildLogs {
  buildId: string
  componentId: string
  orgId: string
}

export async function getBuildLogs({
  orgId,
  componentId,
  buildId,
}: IGetBuildLogs): Promise<TComponentBuildLogs> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}/builds/${buildId}/logs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch build logs')
  }

  return res.json()
}
