import type { TAppRunnerConfig } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetAppRunnerLatestConfig {
  appId: string
  orgId: string
}

export async function getAppRunnerLatestConfig({
  appId,
  orgId,
}: IGetAppRunnerLatestConfig): Promise<TAppRunnerConfig> {
  const data = await fetch(
    `${API_URL}/v1/apps/${appId}/runner-latest-config`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch latest app runner config')
  }

  return data.json()
}
