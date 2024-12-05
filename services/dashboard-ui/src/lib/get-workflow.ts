import type { TActionWorkflow } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetWorkflow {
  orgId: string
  workflowId: string
}

export async function getWorkflow({
  orgId,
  workflowId,
}: IGetWorkflow): Promise<TActionWorkflow> {
  const data = await fetch(
    `${API_URL}/v1/action-workflows/${workflowId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch action workflow')
  }

  return data.json()
}
