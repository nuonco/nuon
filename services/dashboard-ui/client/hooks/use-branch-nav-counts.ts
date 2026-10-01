import { useMemo } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import {
  getAppConfig,
  getAppConfigs,
  getAppLabels,
  getAppPoliciesConfigs,
  getBranchConfigs,
} from '@/lib'

export const useBranchNavCounts = ({
  orgId,
  appId,
  branchId,
}: {
  orgId?: string
  appId?: string
  branchId?: string
}) => {
  const { data: branchConfigs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-configs', orgId, appId, branchId],
    queryFn: () =>
      getBranchConfigs({ orgId: orgId!, appId: appId!, branchId: branchId! }),
    enabled: !!orgId && !!appId && !!branchId,
  })

  const branchAppConfigId = useMemo(
    () =>
      [...(branchConfigs ?? [])].sort(
        (a, b) => (b?.version ?? 0) - (a?.version ?? 0)
      )[0]?.id,
    [branchConfigs]
  )

  const { data: branchAppConfig } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', orgId, appId, branchAppConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: orgId!,
        appId: appId!,
        appConfigId: branchAppConfigId!,
        recurse: true,
      }),
    enabled: !!orgId && !!appId && !!branchAppConfigId,
    retry: false,
  })

  const { data: appConfigs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-configs', orgId, appId],
    queryFn: () => getAppConfigs({ orgId: orgId!, appId: appId!, limit: 1 }),
    enabled: !!orgId && !!appId,
  })

  const latestAppConfigId = appConfigs?.at(0)?.id

  const { data: latestAppConfig } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-config', orgId, appId, latestAppConfigId, 'recurse'],
    queryFn: () =>
      getAppConfig({
        orgId: orgId!,
        appId: appId!,
        appConfigId: latestAppConfigId!,
        recurse: true,
      }),
    enabled: !!orgId && !!appId && !!latestAppConfigId,
  })

  const { data: policiesConfigs } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-policies-configs', orgId, appId],
    queryFn: () => getAppPoliciesConfigs({ orgId: orgId!, appId: appId! }),
    enabled: !!orgId && !!appId,
  })

  const latestPoliciesConfig = useMemo(
    () =>
      policiesConfigs
        ?.slice()
        .sort((a, b) => {
          const dateA = a.created_at ? new Date(a.created_at).getTime() : 0
          const dateB = b.created_at ? new Date(b.created_at).getTime() : 0
          return dateB - dateA
        })
        .at(0),
    [policiesConfigs]
  )

  const { data: labelsData } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-labels', orgId, appId],
    queryFn: () => getAppLabels({ orgId: orgId!, appId: appId! }),
    enabled: !!orgId && !!appId,
  })

  return {
    inputs: branchAppConfig?.input?.inputs?.length,
    roles: latestAppConfig?.permissions?.aws_iam_roles?.length,
    policies: policiesConfigs
      ? (latestPoliciesConfig?.policies?.length ?? 0)
      : undefined,
    labels: labelsData?.labels?.length,
  }
}
