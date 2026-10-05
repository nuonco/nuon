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
    config_repo?: string
    config_directory?: string
  }
}) =>
  api<TApp>({
    path: `apps/${appId}`,
    method: 'PATCH',
    orgId,
    body,
  })
