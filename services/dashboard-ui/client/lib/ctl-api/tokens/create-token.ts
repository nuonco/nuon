import { api } from '@/lib/api'
import type { TCreateTokenBody, TCreateTokenResponse } from '@/types'

export const createToken = ({
  body,
  orgId,
}: {
  body: TCreateTokenBody
  orgId: string
}) =>
  api<TCreateTokenResponse>({
    body,
    method: 'POST',
    orgId,
    path: `tokens`,
  })
