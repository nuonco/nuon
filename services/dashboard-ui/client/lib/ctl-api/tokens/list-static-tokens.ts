import { api } from '@/lib/api'
import type { TStaticToken } from '@/types'

export const listStaticTokens = ({ orgId }: { orgId: string }) =>
  api<TStaticToken[]>({
    path: `account/static-tokens`,
    orgId,
  })
