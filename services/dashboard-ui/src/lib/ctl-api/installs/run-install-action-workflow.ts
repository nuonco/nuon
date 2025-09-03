import { mutateData } from '@/utils'
import type { IGetInstall } from '../shared-interfaces'

export interface IRunInstallActionWorkflow extends IGetInstall {
  actionWorkflowConfigId: string
  options?: {
    run_env_vars: Record<string, string>
  }
}

export async function runInstallActionWorkflow({
  actionWorkflowConfigId,
  installId,
  orgId,
  options,
}: IRunInstallActionWorkflow) {
  return mutateData({
    data: { action_workflow_config_id: actionWorkflowConfigId, ...options },
    errorMessage: 'Unable to run action workflow on this install.',
    orgId,
    path: `installs/${installId}/action-workflows/runs`,
  })
}