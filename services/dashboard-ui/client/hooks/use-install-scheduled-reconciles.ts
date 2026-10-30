import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallScheduledReconciles } from '@/lib'
import type { TAPIError, TInstallScheduledReconcile } from '@/types'

export const INSTALL_SCHEDULED_RECONCILES_QUERY_KEY =
  'install-scheduled-reconciles'

export const useInstallScheduledReconciles = ({
  limit = 20,
}: { limit?: number } = {}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const installId = install?.id

  return useQuery<TInstallScheduledReconcile[], TAPIError>({
    queryKey: [
      INSTALL_SCHEDULED_RECONCILES_QUERY_KEY,
      org?.id,
      installId,
      limit,
    ],
    queryFn: () =>
      getInstallScheduledReconciles({
        orgId: org!.id,
        installId: installId!,
        limit,
      }),
    enabled: !!org?.id && !!installId,
    refetchInterval: 30000,
  })
}

export const usePendingInstallScheduledReconcile = () => {
  const query = useInstallScheduledReconciles()
  const pending = query.data?.find((item) => item.status === 'pending')
  return { ...query, pending }
}
