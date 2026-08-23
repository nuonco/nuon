import { useQuery } from '@tanstack/react-query'
import { getStackServiceAccount } from '@/lib'

// A 404 means the stack has no service account yet, which the caller renders as
// guidance rather than an error — hence retry: false. Not given an infinite
// staleTime like the old token read was: creating a token changes has_live_token,
// so this has to be invalidated and refetched.
export const useStackServiceAccount = ({
  installId,
  orgId,
  enabled,
}: {
  installId?: string
  orgId: string
  enabled: boolean
}) =>
  useQuery({
    queryKey: ['stack-service-account', installId],
    queryFn: () => getStackServiceAccount({ installId: installId!, orgId }),
    enabled: enabled && !!installId,
    retry: false,
  })
