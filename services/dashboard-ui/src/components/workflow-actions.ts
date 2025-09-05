'use server'

import { revalidatePath } from 'next/cache'
import { nueMutateData } from '@/utils'

interface IRunAction {
  actionWorkflowConfigId: string
  orgId: string
  installId: string
  vars: Record<string, string>
}

export async function runAction({
  installId,
  orgId,
  actionWorkflowConfigId,
  vars,
}: IRunAction) {
  const options = Object.keys(vars)?.length ? { run_env_vars: vars } : undefined

  return nueMutateData({
    body: { action_workflow_config_id: actionWorkflowConfigId, ...options },
    orgId,
    path: `installs/${installId}/action-workflows/runs`,
  }).then((r) => {
    revalidatePath(`/${orgId}/installs/${installId}/actions`)
    return r
  })
}

// TODO(nnnat): rename to action workflow history
export async function revalidateInstallWorkflowHistory(
  orgId: string,
  installId: string
) {
  revalidatePath(`/${orgId}/installs/${installId}/actions`)
}

// Install Workflow actions

interface ICancelInstallWorkflow {
  installWorkflowId: string
  orgId: string
}

export async function cancelInstallWorkflow({
  installWorkflowId,
  orgId,
}: ICancelInstallWorkflow) {
  return nueMutateData({
    path: `workflows/${installWorkflowId}/cancel`,
    orgId,
  })
}

interface IInstallWorkflowApproveAll {
  orgId: string
  workflowId: string
}

export async function installWorkflowApproveAll({
  orgId,
  workflowId,
}: IInstallWorkflowApproveAll) {
  return nueMutateData({
    orgId,
    path: `workflows/${workflowId}`,
    method: 'PATCH',
    body: {
      approval_option: 'approve-all',
    },
  })
}
