'use server'

import { revalidatePath } from 'next/cache'
import { deployComponents as deploy, TDeployComponentsBody } from '@/lib'

export async function deployComponents({
  body,
  installId,
  orgId,
  path,
}: {
  body: TDeployComponentsBody
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
