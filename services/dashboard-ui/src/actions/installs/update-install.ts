'use server'

import { revalidatePath } from 'next/cache'
import { updateInstall as update } from '@/lib'

export async function updateInstall({
  installId,
  managedBy,
  orgId,
  path,
}: {
  installId: string
  managedBy: 'nuon/dashboard'
  orgId: string
  path?: string
}) {
  return update({
    body: { metadata: { managed_by: managedBy } },
    installId,
    orgId,
  }).then((res) => {
    if (path) revalidatePath(path)
    return res
  })
}
