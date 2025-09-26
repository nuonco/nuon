'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  createInstallConfig as create,
  type TCreateInstallConfigBody,
} from '@/lib'

export async function createInstallConfig({
  path,
  ...args
}: {
  body: TCreateInstallConfigBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: create,
    args,
    path,
  })
}