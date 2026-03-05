import { api } from '@/lib/api'
import type { TInstallInputs } from '@/types'

export const getInstallCurrentInputs = async ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstallInputs>({
    orgId,
    path: `installs/${installId}/inputs/current`,
  })
