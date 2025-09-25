'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  approveWorkflowStep as approve,
  type TApproveWorkflowStepBody,
} from '@/lib'

export async function approveWorkflowStep({
  path,
  ...args
}: {
  approvalId: string
  body: TApproveWorkflowStepBody
  workflowId: string
  workflowStepId: string
} & IServerAction) {
  return executeServerAction({
    action: approve,
    args,
    path,
  })
}