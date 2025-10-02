'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { removeUser as remove, type TRemoveUserBody } from '@/lib'

export async function removeUser({
  body,
  path,
  ...args
}: {
  body: TRemoveUserBody
} & IServerAction) {
  return executeServerAction({
    action: remove,
    args: { body, ...args },
    path,
  })
}
