'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { removeVCSConnection as removeVCS } from '@/lib'

export async function removeVCSConnection({
  path,
  ...args
}: {
  connectionId: string
} & IServerAction) {
  return executeServerAction({
    action: removeVCS,
    args,
    path,
  })
}
