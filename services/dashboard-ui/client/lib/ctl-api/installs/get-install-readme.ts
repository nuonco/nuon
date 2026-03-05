import { api } from '@/lib/api'
import type { TReadme } from '@/types'

export const getInstallReadme = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TReadme>({
    path: `installs/${installId}/readme`,
    orgId,
  })
