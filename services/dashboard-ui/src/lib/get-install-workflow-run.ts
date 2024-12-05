import type { TInstallActionWorkflowRun } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstallWorkflowRun {
  installId: string
  orgId: string
  workflowRunId: string
}

export async function getInstallWorkflowRun({
  installId,
  orgId,
  workflowRunId,
}: IGetInstallWorkflowRun): Promise<TInstallActionWorkflowRun> {
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}/action-workflows/runs/${workflowRunId}`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch install action workflow run')
  }

  return data.json()
}
