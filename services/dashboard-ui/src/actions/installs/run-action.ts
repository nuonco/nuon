'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { runAction as run, TRunActionBody } from '@/lib'

export async function runAction({
  path,
  ...args
}: {
  body: TRunActionBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: run,
    args,
    path,
  })
}
