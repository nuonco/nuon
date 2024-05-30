import type { TOrg } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetOrg {
  orgId: string
}

export async function getOrg({ orgId }: IGetOrg): Promise<TOrg> {
  const data = await fetch(
    `${API_URL}/v1/orgs/current`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch current org')
  }

  return data.json()
}
