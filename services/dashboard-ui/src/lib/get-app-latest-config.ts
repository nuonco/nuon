import type { TAppConfig } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetAppLatestConfig {
  appId: string
  orgId: string
}

export async function getAppLatestConfig({
  appId,
  orgId,
}: IGetAppLatestConfig): Promise<TAppConfig> {
  const data = await fetch(
    `${API_URL}/v1/apps/${appId}/latest-config`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch latest app config')
  }

  return data.json()
}
