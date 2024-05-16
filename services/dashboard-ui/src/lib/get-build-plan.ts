import type { TComponentBuildPlan } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetBuildPlan {
  buildId: string
  componentId: string
  orgId: string
}

export async function getBuildPlan({
  orgId,
  componentId,
  buildId,
}: IGetBuildPlan): Promise<TComponentBuildPlan> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}/builds/${buildId}/plan`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}
