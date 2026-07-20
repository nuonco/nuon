import { api } from '@/lib/api'

export const deleteStaticToken = ({
  tokenId,
  orgId,
}: {
  tokenId: string
  orgId: string
}) =>
  api<void>({
    method: 'DELETE',
    orgId,
    path: `account/static-tokens/${tokenId}`,
  })
