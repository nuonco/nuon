import type { TApp } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetApps {
  orgId: string
}

export async function getApps({ orgId }: IGetApps): Promise<Array<TApp>> {
  const res = await fetch(`${API_URL}/v1/apps`, await getFetchOpts(orgId))

  if (!res.ok) {
    throw new Error('Failed to fetch apps')
  }

  return res.json()
}
