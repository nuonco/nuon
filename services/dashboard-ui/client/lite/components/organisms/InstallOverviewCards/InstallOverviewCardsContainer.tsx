import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getInstallConfigSyncs } from '@/lib'
import { useInstall } from '../../../providers/install-provider'
import { useOrg } from '../../../providers/org-provider'
import { InstallOverviewCards } from './InstallOverviewCards'

export const InstallOverviewCardsContainer = () => {
  const { orgId } = useOrg()
  const { install, installId, isLoading: isLoadingInstall } = useInstall()
  const { data: syncs, isLoading: isLoadingSyncs } = useQuery({
    queryKey: ['install-config-syncs', orgId, installId],
    queryFn: () =>
      getInstallConfigSyncs({
        orgId: orgId!,
        installId: installId!,
      }),
    enabled: !!orgId && !!installId,
    placeholderData: keepPreviousData,
    refetchInterval: 20_000,
  })

  return (
    <InstallOverviewCards
      install={install}
      latestSync={syncs?.at(0)}
      isLoading={isLoadingInstall || isLoadingSyncs}
    />
  )
}
