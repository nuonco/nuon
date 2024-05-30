import type { TInstall } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstalls {
  orgId: string
}

export async function getInstalls({
  orgId,
}: IGetInstalls): Promise<Array<TInstall>> {
  const data = await fetch(`${API_URL}/v1/installs`, await getFetchOpts(orgId))

  if (!data.ok) {
    throw new Error('Failed to fetch installs')
  }

  return data.json()
}
