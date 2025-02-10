'use server'

import { revalidatePath } from 'next/cache'
import { runInstallActionWorkflow } from '@/lib'

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
  const options =  Object.keys(vars)?.length ? { run_env_vars: vars } : undefined
  
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

export async function revalidateInstallWorkflowHistory(
  orgId: string,
  installId: string
) {
  revalidatePath(`/${orgId}/installs/${installId}/actions`)
}
