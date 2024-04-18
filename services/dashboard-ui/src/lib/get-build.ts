import type { TBuild } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetBuild {
  buildId: string
  componentId: string
  orgId: string
}

export async function getBuild({
  buildId,
  componentId,
  orgId,
}: IGetBuild): Promise<TBuild> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}/builds/${buildId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}
