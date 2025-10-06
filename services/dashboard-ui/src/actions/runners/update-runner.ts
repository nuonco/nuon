'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { updateRunner as update, type IUpdateRunnerBody } from '@/lib'

export async function updateRunner({
  path,
  ...args
}: {
  body: IUpdateRunnerBody
  runnerId: string
} & IServerAction) {
  return executeServerAction({
    action: update,
    args,
    path,
  })
}
