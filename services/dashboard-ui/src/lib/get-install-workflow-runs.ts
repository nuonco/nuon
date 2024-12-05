import type { TInstallActionWorkflowRun } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

export interface IGetInstallWorkflowRuns {
  installId: string
  orgId: string
}

export async function getInstallWorkflowRuns({
  installId,
  orgId,
}: IGetInstallWorkflowRuns): Promise<Array<TInstallActionWorkflowRun>> {
  const data = await fetch(
    `${API_URL}/v1/installs/${installId}/action-workflows/runs`,
    await getFetchOpts(orgId)
  )

  if (!data.ok) {
    throw new Error('Failed to fetch install action workflow runs')
  }

  return data.json()
}
