import type { TBuild } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetComponentBuilds {
  componentId: string
  orgId: string
}

export async function getComponentBuilds({ componentId, orgId }: IGetComponentBuilds): Promise<Array<TBuild>> {
  const res = await fetch(`${API_URL}/v1/builds?component_id=${componentId}`, await getFetchOpts(orgId))

  if (!res.ok) {
    throw new Error('Failed to fetch component builds')
  }

  return res.json()
}
