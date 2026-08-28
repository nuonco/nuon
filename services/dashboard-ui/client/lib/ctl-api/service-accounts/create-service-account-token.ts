import { api } from '@/lib/api'
import type {
  TCreateServiceAccountTokenBody,
  TCreateServiceAccountTokenResponse,
} from '@/types'

export const createServiceAccountToken = ({
  body,
  accountId,
  orgId,
}: {
  body: TCreateServiceAccountTokenBody
  accountId: string
  orgId: string
}) =>
  api<TCreateServiceAccountTokenResponse>({
    body,
    method: 'POST',
    orgId,
    path: `service-accounts/${accountId}/tokens`,
  })
