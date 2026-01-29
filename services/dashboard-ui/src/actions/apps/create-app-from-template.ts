'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  createAppFromTemplate as create,
  type TCreateAppFromTemplateBody,
} from '@/lib'

export async function createAppFromTemplate({
  path,
  ...args
}: {
  body: TCreateAppFromTemplateBody
} & IServerAction) {
  return executeServerAction({
    action: create,
    args,
    path,
  })
}
