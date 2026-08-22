import { api } from '@/lib/api'
import type { TStackToken } from '@/types'

// Returns the API token the install stack's Terraform module authenticates
// with. Read-only: the token is minted during stack version generation, so
// calling this repeatedly never issues a new credential.
export const getStackToken = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) =>
  api<TStackToken>({
    path: `stacks/${installId}/token`,
    orgId,
  })
