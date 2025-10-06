'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { syncSecrets as sync, type TSyncSecretsBody } from '@/lib'

export async function syncSecrets({
  path,
  ...args
}: {
  body: TSyncSecretsBody
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: sync,
    args,
    path,
  })
}
