import type { TBuild } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetBuild {
  buildId: string
  orgId: string
}

export async function getBuild({
  buildId,
  orgId,
}: IGetBuild): Promise<TBuild> {
  const res = await fetch(
    `${API_URL}/v1/components/builds/${buildId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {  
    throw new Error('Failed to fetch build')
  }

  return res.json()
}
