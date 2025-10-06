'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { shutdownRunner as shutdown } from '@/lib'

export async function shutdownRunner({
  path,
  ...args
}: {
  force?: boolean
  runnerId: string
} & IServerAction) {
  return executeServerAction({
    action: shutdown,
    args,
    path,
  })
}
