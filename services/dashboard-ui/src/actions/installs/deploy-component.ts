'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { deployComponent as deploy, TDeployComponentBody } from '@/lib'

export async function deployComponent({
  path,
  ...args
}: {
  body: TDeployComponentBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: deploy,
    args,
    path,
  })
}
