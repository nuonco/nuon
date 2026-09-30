import { api } from '@/lib/api'
import type { TOrg } from '@/types'

export type TUpdateOrgBody = {
  name: string
}

export const updateOrg = ({
  orgId,
  body,
}: {
  orgId: string
  body: TUpdateOrgBody
}) =>
  api<TOrg>({
    body,
    method: 'PATCH',
    orgId,
    path: 'orgs/current',
  })
