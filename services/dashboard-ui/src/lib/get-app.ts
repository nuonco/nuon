import type { TApp } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetApp {
  appId: string
  orgId: string
}

export async function getApp({ appId, orgId }: IGetApp): Promise<TApp> {
  const data = await fetch(
    `${API_URL}/v1/apps/${appId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch app')
  }

  return data.json()
}
