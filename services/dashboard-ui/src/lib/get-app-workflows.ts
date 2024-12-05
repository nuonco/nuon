import type { TActionWorkflow } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetAppWorkflows {
  appId: string
  orgId: string
}

export async function getAppWorkflows({
  appId,
  orgId,
}: IGetAppWorkflows): Promise<Array<TActionWorkflow>> {
  const data = await fetch(
    `${API_URL}/v1/apps/${appId}/action-workflows`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch app action workflows')
  }

  return data.json()
}
