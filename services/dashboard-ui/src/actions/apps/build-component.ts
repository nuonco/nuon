'use server'

import { revalidatePath } from 'next/cache'
import { buildComponent as build } from '@/lib'

export async function buildComponent({
  componentId,
  orgId,
  path,
}: {
  componentId: string
  orgId: string
  path: string
}) {
  return build({ componentId, orgId }).then((res) => {
    revalidatePath(path)
    return res
  })
}
