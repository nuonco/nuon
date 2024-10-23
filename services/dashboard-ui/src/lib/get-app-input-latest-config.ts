import type { TAppInputConfig } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetAppInputLatestConfig {
  appId: string
  orgId: string
}

export async function getAppInputLatestConfig({
  appId,
  orgId,
}: IGetAppInputLatestConfig): Promise<TAppInputConfig> {
  const data = await fetch(
    `${API_URL}/v1/apps/${appId}/input-latest-config`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch latest app input config')
  }

  return data.json()
}
