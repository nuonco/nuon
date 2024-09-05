import type { TInstallDeploy } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstallComponentDeploys {
  installId: string
  installComponentId: string
  orgId: string
}

export async function getInstallComponentDeploys({
  installId,
  installComponentId,
  orgId,
}: IGetInstallComponentDeploys): Promise<Array<TInstallDeploy>> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/components/${installComponentId}/deploys`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch install component deploys')
  }

  return res.json()
}
