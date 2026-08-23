import { api } from '@/lib/api'
import type { TStackServiceAccount } from '@/types'

// Identifies the service account an install stack's Terraform module
// authenticates as, and reports whether it still holds a usable token. Never
// returns a token value — tokens are created through
// createServiceAccountToken, which returns the value exactly once.
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
