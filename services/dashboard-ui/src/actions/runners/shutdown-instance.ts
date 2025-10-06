'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { shutdownRunnerInstance as shutdown } from '@/lib'

export async function shutdownInstance({
  path,
  ...args
}: {
  runnerId: string
} & IServerAction) {
  return executeServerAction({
    action: shutdown,
    args,
    path,
  })
}
