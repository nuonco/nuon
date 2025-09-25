'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  approveAllWorkflowSteps as approveAll,
  type TApproveAllWorkflowStepsBody,
} from '@/lib'

export async function approveAllWorkflowSteps({
  path,
  ...args
}: {
  body: TApproveAllWorkflowStepsBody
  workflowId: string
} & IServerAction) {
  return executeServerAction({
    action: approveAll,
    args,
    path,
  })
}
