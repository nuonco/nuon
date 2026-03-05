import { api } from '@/lib/api'
import type { TInstallStack } from '@/types'

export const getInstallStack = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TInstallStack>({
    path: `installs/${installId}/stack`,
    orgId,
  })
