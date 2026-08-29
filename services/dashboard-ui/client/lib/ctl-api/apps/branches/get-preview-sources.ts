import { api } from '@/lib/api'
import type { TPreviewSources } from '@/types'

export const getPreviewSources = ({
  appId,
  branchId,
  orgId,
}: {
  appId: string
  branchId: string
  orgId: string
}) =>
  api<TPreviewSources>({
    path: `apps/${appId}/branches/${branchId}/preview-sources`,
    orgId,
  })
