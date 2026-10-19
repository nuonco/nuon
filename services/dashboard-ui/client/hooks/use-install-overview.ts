import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallOverview } from '@/lib'
import type { TAPIError, TInstallOverview } from '@/types'

export const INSTALL_OVERVIEW_QUERY_KEY = 'install-overview'

export const useInstallOverview = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const installId = install?.id

  return useQuery<TInstallOverview, TAPIError>({
    queryKey: [
      INSTALL_OVERVIEW_QUERY_KEY,
      org?.id,
      installId,
      install?.app_config_id,
      install?.app_branch?.id ?? install?.app_branch_id,
    ],
    queryFn: () =>
      getInstallOverview({ orgId: org!.id, installId: installId! }),
    enabled: !!org?.id && !!installId,
    refetchInterval: 20000,
    placeholderData: (previous, previousQuery) =>
      previousQuery?.queryKey[1] === org?.id &&
      previousQuery?.queryKey[2] === installId
        ? previous
        : undefined,
  })
}
