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
    // ctl-api only writes these when both are sent.
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
