import type { TInstallDeploy } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstallComponentDeploys {
  componentId: string
  installId: string
  orgId: string
}

export async function getInstallComponentDeploys({
  componentId,
  installId,
  orgId,
}: IGetInstallComponentDeploys): Promise<Array<TInstallDeploy>> {
  const res = await fetch(
    `${API_URL}/v1/installs/${installId}/components/${componentId}/deploys`,
    await getFetchOpts(orgId)
  )

  if (!res.ok) {
    throw new Error('Failed to fetch install component deploys')
  }

  return res.json()
}
