import type { TVCSConnection } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetVCSConnections {
  orgId: string
}

export async function getVCSConnections({
  orgId,
}: IGetVCSConnections): Promise<Array<TVCSConnection>> {
  const res = await fetch(
    `${API_URL}/v1/vcs/connections`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch VCS connections')
  }

  return res.json()
}
