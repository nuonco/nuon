import { api } from '@/lib/api'
import type { TApp } from '@/types'

export const updateApp = ({
  appId,
  orgId,
  body,
}: {
  appId: string
  orgId: string
  body: {
    label_colors?: Record<string, string>
  }
}) =>
  api<TApp>({
    path: `apps/${appId}`,
    method: 'PATCH',
    orgId,
    body,
  })
