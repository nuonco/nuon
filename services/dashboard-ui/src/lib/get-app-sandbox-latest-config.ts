import { IGetAppConfigs } from '@/lib'
import type { TAppSandboxConfig } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export async function getAppSandboxLatestConfig({
  appId,
  orgId,
}: IGetAppConfigs): Promise<TAppSandboxConfig> {
  const data = await fetch(
    `${API_URL}/v1/apps/${appId}/sandbox-latest-config`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch latest app sandbox config')
  }

  return data.json()
}
