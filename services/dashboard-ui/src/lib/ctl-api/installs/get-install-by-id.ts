import { api } from '@/lib/api'
import type { TInstall } from '@/types'

export const getInstallById = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstall>({
    path: `installs/${installId}`,
    orgId,
  })
