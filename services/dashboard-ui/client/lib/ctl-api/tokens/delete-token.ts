import { api } from '@/lib/api'

export const deleteToken = ({
  tokenId,
  orgId,
}: {
  tokenId: string
  orgId: string
}) =>
  api<void>({
    method: 'DELETE',
    orgId,
    path: `tokens/${tokenId}`,
  })
