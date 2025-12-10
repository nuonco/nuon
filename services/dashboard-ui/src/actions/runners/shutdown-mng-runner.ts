'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { shutdownMngRunner as shutdown } from '@/lib'

export async function shutdownMngRunner({
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