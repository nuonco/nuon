import { api } from '@/lib/api'
import type { TStackServiceAccount } from '@/types'

export const getStackServiceAccount = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TStackServiceAccount>({
    path: `stacks/${installId}/service-account`,
    orgId,
  })
