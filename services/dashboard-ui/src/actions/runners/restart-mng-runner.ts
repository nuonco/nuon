'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { restartMngRunner as restart } from '@/lib'

export async function restartMngRunner({
  path,
  ...args
}: {
  runnerId: string
} & IServerAction) {
  return executeServerAction({
    action: restart,
    args,
    path,
  })
}
