import type { TRunner } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetRunner {
  orgId: string
  runnerId: string
}

export async function getRunner({
  orgId,
  runnerId,
}: IGetRunner): Promise<TRunner> {
  const data = await fetch(
    `${API_URL}/v1/runners/${runnerId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch runner')
  }

  return data.json()
}
