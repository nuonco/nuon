import { api } from '@/lib/api'
import type { TComponentConfig } from '@/types'

export const getComponentConfig = ({
  appId,
  componentId,
  configId,
  orgId,
}: {
  appId: string
  componentId: string
  configId: string
  orgId: string
}) =>
  api<TComponentConfig>({
    path: `apps/${appId}/components/${componentId}/configs/${configId}`,
    orgId,
  })