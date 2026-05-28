import { api } from '@/lib/api'
import type { TRunbook } from './get-runbooks'

export const getRunbook = ({
  runbookId,
  appId,
  orgId,
}: {
  runbookId: string
  appId: string
  orgId: string
}) =>
  api<TRunbook>({
    path: `apps/${appId}/runbooks/${runbookId}`,
    orgId,
  })
