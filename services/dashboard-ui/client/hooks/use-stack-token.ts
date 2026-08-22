import { useQuery } from '@tanstack/react-query'
import { getStackToken } from '@/lib'

// The token is long-lived and only ever read, so there is nothing to refetch
// for. A 404 means the stack has not minted one yet, which the caller renders
// as guidance rather than an error.
export const useStackToken = ({
  installId,
  orgId,
  enabled,
}: {
  installId?: string
  orgId: string
  enabled: boolean
}) =>
  useQuery({
    queryKey: ['stack-token', installId],
    queryFn: () => getStackToken({ installId: installId!, orgId }),
    enabled: enabled && !!installId,
    staleTime: Infinity,
    retry: false,
  })
