import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'
import { getInstallReconciliation } from '@/lib'
import type { TAPIError, TInstallReconciliation } from '@/types'

export const INSTALL_RECONCILIATION_FLAG = 'install-reconciliation'
export const INSTALL_RECONCILIATION_QUERY_KEY = 'install-reconciliation'

export const useInstallReconciliation = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const enabled = useOrgFeatureFlag(INSTALL_RECONCILIATION_FLAG)

  return useQuery<TInstallReconciliation, TAPIError>({
    queryKey: [
      INSTALL_RECONCILIATION_QUERY_KEY,
      org?.id,
      install?.id,
      install?.app_config_id,
      install?.app_branch?.id ?? install?.app_branch_id,
    ],
    queryFn: () =>
      getInstallReconciliation({ orgId: org!.id, installId: install!.id }),
    enabled: enabled && !!org?.id && !!install?.id,
    refetchInterval: 20000,
    placeholderData: keepPreviousData,
  })
}
