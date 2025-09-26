'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  deprovisionSandbox as deprovision,
  type TDeprovisionSandboxBody,
} from '@/lib'

export async function deprovisionSandbox({
  path,
  ...args
}: {
  body: TDeprovisionSandboxBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: deprovision,
    args,
    path,
  })
}