'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { inviteUser as invite, type TInviteUserBody } from '@/lib'

export async function inviteUser({
  body,
  path,
  ...args
}: {
  body: TInviteUserBody
} & IServerAction) {
  return executeServerAction({
    action: invite,
    args: { body, ...args },
    path,
  })
}
