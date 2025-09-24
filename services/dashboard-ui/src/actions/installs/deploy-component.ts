'use server'

import { revalidatePath } from 'next/cache'
import { deployComponent as deploy, TDeployComponentBody } from '@/lib'

export async function deployComponent({
  body,
  installId,
  orgId,
  path,
}: {
  body: TDeployComponentBody
  installId: string
  orgId: string
  path?: string
}) {
  return deploy({
    body,
    installId,
    orgId,
  }).then((res) => {
    if (path) revalidatePath(path)
    return res
  })
}
