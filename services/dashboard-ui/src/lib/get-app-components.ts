import type { TComponent } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetAppComponents {
  appId: string
  orgId: string
}

export async function getAppComponents({ appId, orgId }: IGetAppComponents): Promise<Array<TComponent>> {
  const res = await fetch(`${API_URL}/v1/apps/${appId}/components`, await getFetchOpts(orgId))

  if (!res.ok) {
    throw new Error('Failed to fetch app components')
  }

  return res.json()
}
