'use server'

import { revalidatePath } from 'next/cache'
import { postWorkflowRun } from '@/lib'

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
    postWorkflowRun({ installId, orgId, workflowConfigId })
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
