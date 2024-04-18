import type { TComponent } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetComponent {
  componentId: string
  orgId: string
}

export async function getComponent({
  componentId,
  orgId,
}: IGetComponent): Promise<TComponent> {
  const res = await fetch(
    `${API_URL}/v1/components/${componentId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}
