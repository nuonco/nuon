'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { forgetComponent as forget } from '@/lib'

export async function forgetComponent({
  path,
  ...args
}: {
  componentId: string
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: forget,
    args,
    path,
  })
}
