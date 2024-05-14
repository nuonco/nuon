import type { TOrg } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export async function postOrg(data: Record<string, string>): Promise<TOrg> {
  const res = await fetch(`${API_URL}/v1/orgs`, {
    ...(await getFetchOpts()),
    body: JSON.stringify(data),
    method: 'POST',
  })

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}
