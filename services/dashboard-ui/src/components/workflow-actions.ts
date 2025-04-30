'use server'

import { revalidatePath } from 'next/cache'
import { runInstallActionWorkflow } from '@/lib'
import { nueMutateData } from '@/utils'

// TODO(nnnat): rename these to run action workflows
interface IRunManualWorkflow {
  orgId: string
  installId: string
  workflowConfigId: string
  vars: Record<string, string>
}

export async function runManualWorkflow({
  installId,
  orgId,
  workflowConfigId,
  vars,
}: IRunManualWorkflow) {
  const options = Object.keys(vars)?.length ? { run_env_vars: vars } : undefined

  try {
    await runInstallActionWorkflow({
      installId,
      orgId,
      actionWorkflowConfigId: workflowConfigId,
      options,
    })
    revalidatePath(`/${orgId}/installs/${installId}/actions`)
  } catch (error) {
    throw new Error(error.message)
  }
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
    path: `install-workflows/${installWorkflowId}/cancel`,
    orgId,
  })
}
