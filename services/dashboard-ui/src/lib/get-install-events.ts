import type { TInstallEvent } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstallEvents {
  installId: string
  orgId: string
}

export async function getInstallEvents({
  installId,
  orgId,
}: IGetInstallEvents): Promise<Array<TInstallEvent>> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/events`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch install events')
  }

  return res.json()
}
