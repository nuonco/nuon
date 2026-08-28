import { useQuery } from '@tanstack/react-query'
import { getStackServiceAccount } from '@/lib'

// A 404 means no service account yet, which the caller renders as guidance — hence
// retry: false.
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
