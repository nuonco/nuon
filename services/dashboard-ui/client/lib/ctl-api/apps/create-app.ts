import { api } from '@/lib/api'
import type { TApp } from '@/types'

export const createApp = ({
  orgId,
  body,
}: {
  orgId: string
  body: { name: string }
}) =>
  api<TApp>({
    path: 'apps',
    method: 'POST',
    orgId,
    body,
  })
