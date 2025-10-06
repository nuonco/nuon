'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { forgetInstall as forget } from '@/lib'

export async function forgetInstall({
  path,
  ...args
}: {
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: forget,
    args,
    path,
  })
}
