'use server'

import { revalidatePath } from 'next/cache'
import { updateInstallInputs as update } from '@/lib'

export async function updateInstallInputs({
  installId,
  formData: fd,
  orgId,
  path,
}: {
  installId: string
  formData: FormData
  orgId: string
  path: string
}) {
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

  return update({ installId, orgId, body: { inputs } }).then((res) => {
    revalidatePath(path)
    return { ...res, headers: Object.fromEntries(res.headers) }
  })
}
