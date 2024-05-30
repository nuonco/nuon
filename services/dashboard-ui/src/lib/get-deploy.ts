import type { TInstallDeploy } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetDeploy {
  deployId: string
  installId: string
  orgId: string
}

export async function getDeploy({
  orgId,
  installId,
  deployId,
}: IGetDeploy): Promise<TInstallDeploy> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch deploy')
  }

  return res.json()
}
