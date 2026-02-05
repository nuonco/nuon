'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { pruneRunnerTokens as prune } from '@/lib'

export async function pruneRunnerTokens({
  path,
  ...args
}: {
  runnerId: string
} & IServerAction) {
  return executeServerAction({
    action: prune,
    args,
    path,
  })
}
