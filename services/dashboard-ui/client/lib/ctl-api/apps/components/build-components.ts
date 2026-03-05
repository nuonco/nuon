import { api } from '@/lib/api'
import type { TBuild } from '@/types'

export const buildComponents = ({
  appId,
  orgId,
}: {
  appId: string
  orgId: string
}) =>
  api<TBuild[]>({
    path: `apps/${appId}/components/build-all`,
    method: 'POST',
    orgId,
    body: {},
  })
