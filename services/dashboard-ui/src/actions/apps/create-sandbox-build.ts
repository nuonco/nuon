'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { createSandboxBuild as create } from '@/lib'

export async function createSandboxBuild({
  path,
  ...args
}: {
  appId: string
} & IServerAction) {
  return executeServerAction({
    action: create,
    args,
    path,
  })
}
