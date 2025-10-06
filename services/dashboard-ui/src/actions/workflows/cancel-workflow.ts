'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { cancelWorkflow as cancel } from '@/lib'

export async function cancelWorkflow({
  path,
  ...args
}: {
  workflowId: string
} & IServerAction) {
  return executeServerAction({
    action: cancel,
    args,
    path,
  })
}
