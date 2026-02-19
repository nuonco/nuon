'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { resendOrgInvite as resend } from '@/lib'

export async function resendOrgInvite({
  inviteId,
  path,
  ...args
}: {
  inviteId: string
} & IServerAction) {
  return executeServerAction({
    action: resend,
    args: { inviteId, ...args },
    path,
  })
}
