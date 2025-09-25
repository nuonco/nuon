'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { buildComponent as build } from '@/lib'

export async function buildComponent({
  path,
  ...args
}: {
  componentId: string
} & IServerAction) {
  return executeServerAction({
    action: build,
    args,
    path,
  })
}
