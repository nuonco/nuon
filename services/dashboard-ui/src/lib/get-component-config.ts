import type { TComponentConfig } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetComponentConfig {
  componentId: string
  orgId: string
}

export async function getComponentConfig({
  componentId,
  orgId,
}: IGetComponentConfig): Promise<TComponentConfig> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}/configs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json().then((cfgs) => cfgs?.[0])
}
