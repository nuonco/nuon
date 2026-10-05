import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getAppConfig, getAppConfigs, getBranchConfigs } from '@/lib'

export const useBranchScopedAppConfig = ({
  orgId,
  appId,
  branchId,
}: {
  orgId?: string
  appId?: string
  branchId?: string
}) => {
  const scopedToBranch = !!branchId

  const {
    data: branchConfigs,
    isFetched: branchConfigsFetched,
    isError: isBranchConfigsError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-configs', orgId, appId, branchId],
    queryFn: () =>
      getBranchConfigs({ orgId: orgId!, appId: appId!, branchId: branchId! }),
    enabled: !!orgId && !!appId && scopedToBranch,
  })

  const { data: appConfigs, isFetched: appConfigsFetched } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', orgId, appId],
    queryFn: () => getAppConfigs({ orgId: orgId!, appId: appId!, limit: 1 }),
    enabled: !!orgId && !!appId && !scopedToBranch,
  })

  const appConfigId = useMemo(() => {
    if (scopedToBranch) {
      return [...(branchConfigs ?? [])].sort(
        (a, b) => (b?.version ?? 0) - (a?.version ?? 0)
      )[0]?.id
    }
    return appConfigs?.at(0)?.id
  }, [scopedToBranch, branchConfigs, appConfigs])

  const {
    data: appConfig,
    isLoading: isLoadingConfig,
    isFetched: configFetched,
    isError: isConfigError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', orgId, appId, appConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: orgId!,
        appId: appId!,
        appConfigId: appConfigId!,
        recurse: true,
      }),
    enabled: !!orgId && !!appId && !!appConfigId,
    retry: scopedToBranch ? false : undefined,
  })

  const listFetched = scopedToBranch ? branchConfigsFetched : appConfigsFetched
  const isLoading =
    !listFetched || (!!appConfigId && isLoadingConfig && !configFetched)

  return {
    appConfig,
    isLoading,
    isError: isBranchConfigsError || isConfigError,
  }
}
