'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  deprovisionInstall as deprovision,
  type TDeprovisionInstallBody,
} from '@/lib'

export async function deprovisionInstall({
  path,
  ...args
}: {
  body: TDeprovisionInstallBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: deprovision,
    args,
    path,
  })
}
