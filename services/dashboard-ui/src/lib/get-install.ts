import type { TInstall } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstall {
  installId: string
  orgId: string
}

export async function getInstall({
  installId,
  orgId,
}: IGetInstall): Promise<TInstall> {
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch data')
  }

  return data.json()
}
