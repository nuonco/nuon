import type { TOrg } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export async function getOrgs(): Promise<Array<TOrg>> {
  const res = await fetch(`${API_URL}/v1/orgs`, await getFetchOpts())

  if (!res.ok) {
    throw new Error('Failed to fetch orgs')
  }

  return res.json()
}
