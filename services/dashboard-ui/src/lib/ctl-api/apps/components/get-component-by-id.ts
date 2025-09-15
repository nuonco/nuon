import { api } from '@/lib/api'
import type { TComponent } from '@/types'

export const getComponentById = ({
  componentId,
  orgId,
}: {
  componentId: string
  orgId: string
}) =>
  api<TComponent>({
    path: `components/${componentId}`,
    orgId,
  })
