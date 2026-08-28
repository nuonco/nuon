import { api } from '@/lib/api'
import type {
  TCreateStaticTokenBody,
  TCreateStaticTokenResponse,
} from '@/types'

export const createStaticToken = ({
  body,
  orgId,
}: {
  body: TCreateStaticTokenBody
  orgId: string
}) =>
  api<TCreateStaticTokenResponse>({
    body,
    method: 'POST',
    orgId,
    path: `account/static-token`,
  })
