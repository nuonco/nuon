'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { deployComponents as deploy, TDeployComponentsBody } from '@/lib'

export async function deployComponents({
  path,
  ...args
}: {
  body: TDeployComponentsBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: deploy,
    args,
    path,
  })
}
