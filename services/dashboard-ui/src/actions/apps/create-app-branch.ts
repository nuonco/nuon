'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { createAppBranch as create, type TCreateAppBranchRequest } from '@/lib'

export async function createAppBranch({
  body,
  path,
  ...args
}: {
  appId: string
  body: TCreateAppBranchRequest
} & IServerAction) {
  return executeServerAction({
    action: create,
    args: { ...args, body },
    path,
  })
}
