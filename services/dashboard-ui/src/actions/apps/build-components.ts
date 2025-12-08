'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { buildComponents as build } from '@/lib'

export async function buildComponents({
  path,
  ...args
}: {
  appId: string
} & IServerAction) {
  return executeServerAction({
    action: build,
    args,
    path,
  })
}
