'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import {
  updateInstallConfig as update,
  type TUpdateInstallConfigBody,
} from '@/lib/ctl-api/installs/update-install-config'

export async function updateInstallConfig({
  path,
  ...args
}: {
  body: TUpdateInstallConfigBody
  installConfigId: string
  installId: string
} & IServerAction) {
  return executeServerAction({
    action: update,
    args,
    path,
  })
}
