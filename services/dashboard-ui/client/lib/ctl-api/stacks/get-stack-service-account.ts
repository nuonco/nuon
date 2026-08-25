import { api } from '@/lib/api'
import type { TStackServiceAccount } from '@/types'

// Identifies the service account an install stack authenticates as, and whether it
// holds a usable token. Never the token value: createServiceAccountToken returns
// that exactly once.
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
