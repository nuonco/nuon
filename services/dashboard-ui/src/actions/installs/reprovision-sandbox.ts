'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  reprovisionSandbox as reprovision,
  type TReprovisionSandboxBody,
} from '@/lib'

export async function reprovisionSandbox({
  path,
  ...args
}: {
  body: TReprovisionSandboxBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: reprovision,
    args,
    path,
  })
}
