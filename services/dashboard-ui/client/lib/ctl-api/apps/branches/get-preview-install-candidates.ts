import { api } from '@/lib/api'
import type { TInstall } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getPreviewInstallCandidates = ({
  appId,
  branchId,
  orgId,
  configId,
}: {
  appId: string
  branchId: string
  orgId: string
  configId?: string
}) =>
  api<{ installs: TInstall[] }>({
    path: `apps/${appId}/branches/${branchId}/preview-install-candidates${buildQueryParams({ config_id: configId })}`,
    orgId,
  })
