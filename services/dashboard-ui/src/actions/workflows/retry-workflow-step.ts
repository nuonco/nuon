'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { retryWorkflowStep as retry, type TRetryWorkflowStepBody } from '@/lib'

export async function retryWorkflowStep({
  path,
  ...args
}: {
  body: TRetryWorkflowStepBody
  workflowId: string
} & IServerAction) {
  return executeServerAction({
    action: retry,
    args,
    path,
  })
}
