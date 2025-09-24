'use server'

import { revalidatePath } from 'next/cache'
import { teardownComponents as teardown, TTeardownComponentsBody } from '@/lib'

export async function teardownComponents({
  body,
  installId,
  orgId,
  path,
}: {
  body: TTeardownComponentsBody
  installId: string
  orgId: string
  path?: string
}) {
  return teardown({
    body,
    installId,
    orgId,
  }).then((res) => {
    if (path) revalidatePath(path)
    return res
  })
}
