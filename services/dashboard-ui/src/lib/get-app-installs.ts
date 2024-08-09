import type { TInstall } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetAppInstalls {
  appId: string
  orgId: string
}

export async function getAppInstalls({ appId, orgId }: IGetAppInstalls): Promise<Array<TInstall>> {
  const res = await fetch(`${API_URL}/v1/apps/${appId}/installs`, await getFetchOpts(orgId))

  if (!res.ok) {
    throw new Error('Failed to fetch app installs')
  }

  return res.json()
}
