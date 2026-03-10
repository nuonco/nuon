import { api } from '@/lib/api'
import type { TFileResponse } from '@/types'

export const generateCLIInstallConfig = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TFileResponse>({
    path: `installs/${installId}/generate-cli-install-config`,
    orgId,
  })
