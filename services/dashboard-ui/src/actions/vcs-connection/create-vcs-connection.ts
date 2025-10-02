'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { createVCSConnection as create } from '@/lib'

type TCreateVCSConnectionBody = {
  github_install_id: string
}

export async function createVCSConnection({
  body,
  path,
  ...args
}: {
  body: TCreateVCSConnectionBody
} & IServerAction) {
  return executeServerAction({
    action: create,
    args: { body, ...args },
    path,
  })
}
