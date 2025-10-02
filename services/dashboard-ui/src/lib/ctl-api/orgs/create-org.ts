import { api } from '@/lib/api'
import type { TOrg } from '@/types'

export type TCreateOrgBody = {
  name: string
  use_sandbox_mode: boolean
}

export const createOrg = ({ body }: { body: TCreateOrgBody }) =>
  api<TOrg>({
    body,
    method: 'POST',
    path: `orgs`,
  })
