import type { TRunnerGroup } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstallRunnerGroup {
  installId: string
  orgId: string
}

export async function getInstallRunnerGroup({
  installId,
  orgId,
}: IGetInstallRunnerGroup): Promise<TRunnerGroup> {
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}/runner-group`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch install runner group')
  }

  return data.json()
}
