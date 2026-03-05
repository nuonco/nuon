'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { createRunnerBootstrapToken as create } from '@/lib'

export async function createRunnerBootstrapToken({
  path,
  ...args
}: {
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: create,
    args,
    path,
  })
}
