import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useBranch } from '@/hooks/use-branch'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getBranchConfigs } from '@/lib'
import { BranchInputs } from './BranchInputs'

export const BranchInputsContainer = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()

  const {
    data: configs,
    isLoading: isLoadingConfigs,
    isError: isConfigsError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-configs', org?.id, app?.id, branch?.id],
    queryFn: () =>
      getBranchConfigs({
        orgId: org!.id,
        appId: app!.id,
        branchId: branch.id!,
      }),
    enabled: !!org?.id && !!app?.id && !!branch?.id,
  })

  const appConfigId = useMemo(
    () =>
      [...(configs ?? [])].sort(
        (a, b) => (b?.version ?? 0) - (a?.version ?? 0)
      )[0]?.id,
    [configs]
  )

  const {
    data: appConfig,
    isLoading: isLoadingConfig,
    isError: isConfigError,
  } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', org?.id, app?.id, appConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: org!.id,
        appId: app!.id,
        appConfigId: appConfigId!,
        recurse: true,
      }),
    enabled: !!org?.id && !!app?.id && !!appConfigId,
  })

  return (
    <BranchInputs
      appConfig={appConfig}
      isLoading={isLoadingConfigs || (!!appConfigId && isLoadingConfig)}
      isError={isConfigsError || isConfigError}
    />
  )
}
