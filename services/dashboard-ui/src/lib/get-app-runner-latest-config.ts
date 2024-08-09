import { IGetAppConfigs } from '@/lib'
import type { TAppRunnerConfig } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export async function getAppRunnerLatestConfig({
  appId,
  orgId,
}: IGetAppConfigs): Promise<TAppRunnerConfig> {
  const data = await fetch(
    `${API_URL}/v1/apps/${appId}/runner-latest-config`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch latest app runner config')
  }

  return data.json()
}
