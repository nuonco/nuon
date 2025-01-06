'use server'

import { revalidatePath } from 'next/cache'
import { runInstallActionWorkflow } from '@/lib'

interface IRunManualWorkflow {
  orgId: string
  installId: string
  workflowConfigId: string
}

export async function runManualWorkflow({
  installId,
  orgId,
  workflowConfigId,
}: IRunManualWorkflow) {
  try {
    runInstallActionWorkflow({
      installId,
      orgId,
      actionWorkflowConfigId: workflowConfigId,
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
