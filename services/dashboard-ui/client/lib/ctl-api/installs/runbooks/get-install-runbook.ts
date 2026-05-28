import { api } from '@/lib/api'
import type { TInstallRunbook } from './get-install-runbooks'

export const getInstallRunbook = ({
  installId,
  runbookId,
  orgId,
}: {
  installId: string
  runbookId: string
  orgId: string
}) =>
  api<TInstallRunbook>({
    path: `installs/${installId}/runbooks/${runbookId}`,
    orgId,
  })
