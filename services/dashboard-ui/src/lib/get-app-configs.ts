import type { TAppConfig } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetAppConfigs {
  appId: string
  orgId: string
}

export async function getAppConfigs({
  appId,
  orgId,
}: IGetAppConfigs): Promise<Array<TAppConfig>> {
  const data = await fetch(
    `${API_URL}/v1/apps/${appId}/configs`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch app configs')
  }

  return data.json()
}
