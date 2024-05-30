import type { TComponentConfig } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetComponentConfig {
  componentId: string
  componentConfigId?: string
  orgId: string
}

export async function getComponentConfig({
  componentId,
  componentConfigId = 'latest',
  orgId,
}: IGetComponentConfig): Promise<TComponentConfig> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}/configs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch component config')
  }

  return res
    .json()
    .then((cfgs) =>
      componentConfigId === 'latest'
        ? cfgs?.[0]
        : cfgs?.find((cfg) => cfg.id === componentConfigId)
    )
}
