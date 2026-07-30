import { api } from '@/lib/api'
import type { TConfigDiffNode } from '@/types'

export const getInstallConfigVersionDiff = ({
  installId,
  versionId,
  orgId,
}: {
  installId: string
  versionId: string
  orgId: string
}) =>
  api<TConfigDiffNode>({
    path: `installs/${installId}/config-versions/${versionId}/diff`,
    orgId,
  })
