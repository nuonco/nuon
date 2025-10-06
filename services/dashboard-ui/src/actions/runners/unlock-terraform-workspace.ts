'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { unlockTerraformWorkspace as unlock } from '@/lib'

export async function unlockTerraformWorkspace({
  path,
  ...args
}: {
  terraformWorkspaceId: string
} & IServerAction) {
  return executeServerAction({
    action: unlock,
    args,
    path,
  })
}
