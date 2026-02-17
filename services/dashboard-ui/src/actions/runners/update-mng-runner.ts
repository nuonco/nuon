'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { updateMngRunner as update } from '@/lib'

export async function updateMngRunner({
  path,
  ...args
}: {
  runnerId: string
} & IServerAction) {
  return executeServerAction({
    action: update,
    args,
    path,
  })
}
