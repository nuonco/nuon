import { api } from '@/lib/api'

export const deleteServiceAccount = ({
  accountId,
  orgId,
}: {
  accountId: string
  orgId: string
}) =>
  api<void>({
    method: 'DELETE',
    orgId,
    path: `service-accounts/${accountId}`,
  })
