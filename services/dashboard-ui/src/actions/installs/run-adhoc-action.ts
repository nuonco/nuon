'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { runAdhocAction as run, type TRunAdhocActionBody } from '@/lib'

export async function runAdhocAction({
  path,
  ...args
}: {
  body: TRunAdhocActionBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: run,
    args,
    path,
  })
}
