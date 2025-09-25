'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { updateInstall as update, TUpdateInstallBody } from '@/lib'

export async function updateInstall({
  path,
  ...args
}: {
  body: TUpdateInstallBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: update,
    args,
    path,
  })
}
