import { useMemo } from 'react'
import { keepPreviousData, useQuery, useQueries } from '@tanstack/react-query'
import { getAppConfig, getBranchWorkflowRuns } from '@/lib'
import {
  lastApprovedGroupVersions,
  type TGroupApprovedVersion,
} from './group-approved-version'

export const useGroupApprovedVersions = ({
  orgId,
  appId,
  branchId,
  repoSlug,
  enabled = true,
}: {
  orgId?: string
  appId?: string
  branchId?: string
  repoSlug?: string
  enabled?: boolean
}) => {
  const runs = useQuery({
    queryKey: ['branch-approved-group-versions', orgId, appId, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId: orgId!,
        appId: appId!,
        branchId: branchId!,
        limit: 40,
        offset: 0,
        preview: false,
      }),
    enabled: enabled && !!orgId && !!appId && !!branchId,
    refetchInterval: 5000,
    placeholderData: keepPreviousData,
  })

  const base = useMemo(
    () => lastApprovedGroupVersions(runs.data?.data ?? [], repoSlug),
    [repoSlug, runs.data?.data]
  )
  const configIds = useMemo(
    () => [
      ...new Set(
        Object.values(base)
          .map((version) => version.configId)
          .filter((id): id is string => !!id)
      ),
    ],
    [base]
  )
  const configs = useQueries({
    queries: configIds.map((configId) => ({
      queryKey: ['app-config', orgId, appId, configId],
      queryFn: () =>
        getAppConfig({
          orgId: orgId!,
          appId: appId!,
          appConfigId: configId,
        }),
      enabled: !!orgId && !!appId,
      staleTime: Infinity,
    })),
  })
  const versionByConfig = new Map(
    configIds.map((id, index) => [id, configs[index]?.data?.version] as const)
  )
  const versions: Record<string, TGroupApprovedVersion> = Object.fromEntries(
    Object.entries(base).map(([name, version]) => [
      name,
      {
        ...version,
        version: version.configId
          ? versionByConfig.get(version.configId)
          : undefined,
      },
    ])
  )

  return {
    versions,
    isLoading: enabled && runs.isLoading && !runs.data,
  }
}
