'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  reprovisionInstall as reprovision,
  type TReprovisionInstallBody,
} from '@/lib'

export async function reprovisionInstall({
  path,
  ...args
}: {
  body: TReprovisionInstallBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: reprovision,
    args,
    path,
  })
}