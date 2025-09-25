'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { buildComponents as build } from '@/lib'
import type { TComponent } from '@/types'

// NOTE: The buildComponents lib function is special as
// it maps over each component then creates a build.
// This means the return type doesn't conform to the standard
// TAPIResponse type and can not be used with the useServerAction hook
export async function buildComponents({
  path,
  ...args
}: {
  components: TComponent[]
} & IServerAction) {
  return executeServerAction({
    action: build,
    args,
    path,
  })
}
