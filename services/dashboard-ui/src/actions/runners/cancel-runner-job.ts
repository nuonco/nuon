'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { cancelRunnerJob as cancel } from '@/lib'

export async function cancelRunnerJob({
  path,
  ...args
}: {
  runnerJobId: string
} & IServerAction) {
  return executeServerAction({
    action: cancel,
    args,
    path,
  })
}
