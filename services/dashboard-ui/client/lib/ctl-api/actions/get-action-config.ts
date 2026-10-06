import { api } from '@/lib/api'
import type { TActionConfig } from '@/types'

export const getActionConfig = ({
  actionConfigId,
  orgId,
}: {
  actionConfigId: string
  orgId: string
}) =>
  api<TActionConfig>({
    path: `action-workflows/configs/${actionConfigId}`,
    orgId,
  })
