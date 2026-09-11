import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getInstallConfigSyncs } from '@/lib'
import { useInstall } from '../../../providers/install-provider'
import { useOrg } from '../../../providers/org-provider'
import { appBranchActivityHref } from '../../../utils/hrefs'
import { InstallOverviewCards } from './InstallOverviewCards'

export const InstallOverviewCardsContainer = () => {
  const { orgId } = useOrg()
  const { install, installId, loading: installLoading } = useInstall()
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
  const sync = syncs?.at(0)

  return (
    <InstallOverviewCards
      install={install}
      lastBranchUpdate={
        sync
          ? {
              runId: sync.app_branch_run_id,
              runHref:
                sync.app_branch_run_id &&
                orgId &&
                install?.app_id &&
                install?.app_branch_id
                ? appBranchActivityHref(
                    orgId,
                    install.app_id,
                    install.app_branch_id
                  )
                : undefined,
              branchName: install?.app_branch?.name,
              status: sync.status?.status,
              commit: sync.vcs_connection_commit,
              updatedAt: sync.created_at,
            }
          : undefined
      }
      loading={installLoading || isLoadingSyncs}
    />
  )
}
