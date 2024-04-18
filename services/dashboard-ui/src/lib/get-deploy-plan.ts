import type { TInstallDeployPlan } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetDeployPlan {
  deployId: string
  installId: string
  orgId: string
}

export async function getDeployPlan({
  orgId,
  installId,
  deployId,
}: IGetDeployPlan): Promise<TInstallDeployPlan> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/deploys/${deployId}/plan`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}
