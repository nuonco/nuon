'use server'

import { revalidatePath } from 'next/cache'
import { buildComponents as build } from '@/lib'
import type { TComponent } from '@/types'

// NOTE: The buildComponents lib function is special as
// it maps over each component then creates a build.
// This means the return type doesn't conform to the standard
// TAPIResponse type and can not be used with the useServerAction hook
export async function buildComponents({
  components,
  orgId,
  path,
}: {
  components: TComponent[]
  orgId: string
  path: string
}) {
  return build({ components, orgId }).then((res) => {
    revalidatePath(path)
    return res
  })
}
