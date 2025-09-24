'use server'

import { revalidatePath } from 'next/cache'
import { teardownComponent as teardown,  TTeardownComponentBody } from '@/lib'

export async function teardownComponent({
  body,
  componentId,
  installId,
  orgId,
  path,
}: {
  body: TTeardownComponentBody
  componentId: string
  installId: string
  orgId: string
  path?: string
}) {
  return teardown({
    body,
    componentId,
    installId,
    orgId,
  }).then((res) => {
    if (path) revalidatePath(path)
    return res
  })
}
