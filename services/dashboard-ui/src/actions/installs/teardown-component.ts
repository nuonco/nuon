'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { teardownComponent as teardown,  TTeardownComponentBody } from '@/lib'

export async function teardownComponent({
  path,
  ...args
}: {
  body: TTeardownComponentBody
  componentId: string
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: teardown,
    args,
    path,
  })
}
