import type { TInstallDeployLogs } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetDeployLogs {
  deployId: string
  installId: string
  orgId: string
}

export async function getDeployLogs({
  orgId,
  installId,
  deployId,
}: IGetDeployLogs): Promise<TInstallDeployLogs> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}/logs`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}
