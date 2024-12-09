import type { TInstallActionWorkflowRun } from '@/types'
import { API_URL, getFetchOpts } from '@/utils'

interface IPostWorkflowRun {
  orgId: string
  installId: string 
  workflowConfigId: string
}

export async function postWorkflowRun({ installId, orgId, workflowConfigId }: IPostWorkflowRun): Promise<TInstallActionWorkflowRun> {
  const res = await fetch(`${API_URL}/v1/installs/${installId}/action-workflows/runs`, { 
    ...(await getFetchOpts(orgId)),
    body: JSON.stringify({ action_workflow_config_id: workflowConfigId }),
    method: 'POST',
  })

  if (!res.ok) {
    throw new Error('Failed to kick off an action workflow')
  }

  return res.json()
}
