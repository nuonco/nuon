'use server'

import {
  executeServerAction,
  type IServerAction,
} from '@/actions/execute-server-action'
import { updateInstallInputs as update } from '@/lib'

export async function updateInstallInputs({
  formData: fd,
  path,
  ...args
}: {
  installId: string
  formData: FormData
} & IServerAction) {
  const formData = Object.fromEntries(fd)
  const inputs = Object.keys(formData).reduce((acc, key) => {
    if (key.includes('inputs:')) {
      let value: any = formData[key]
      if (value === 'on' || value === 'off') {
        value = Boolean(value === 'on').toString()
      }

      acc[key.replace('inputs:', '')] = value
    }

    return acc
  }, {})

  return executeServerAction({
    action: update,
    args: { ...args, body: { inputs } },
    path,
  })
}
