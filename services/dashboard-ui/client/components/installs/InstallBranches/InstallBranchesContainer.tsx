import { useMemo } from 'react'
import { useQueries, useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getBranchWorkflowRuns, getBranchRunBuilds, getBranchRunInstallGroups, getInstallAppConfigVersions } from '@/lib'
import { InstallBranches, type IBranchEntry } from './InstallBranches'
import type { TInstall } from '@/types'

interface IInstallBranchesContainer {
  install?: TInstall
}

export const InstallBranchesSection = ({ install }: IInstallBranchesContainer) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''
  const appId = install?.app_id ?? ''

  const connections = useMemo(() => {
    const conns = install?.app_branch_connections ?? []
    const seen = new Set<string>()
    return conns.filter((c) => {
      const bid = c.app_branch_id
      if (!bid || seen.has(bid)) return false
      seen.add(bid)
      return true
    })
  }, [install?.app_branch_connections])

  const runQueries = useQueries({
    queries: connections.map((conn) => ({
      queryKey: ['branch-latest-run', orgId, appId, conn.app_branch_id],
      queryFn: () =>
        getBranchWorkflowRuns({
          orgId: orgId!,
          appId: appId!,
          branchId: conn.app_branch_id!,
          limit: 1,
          offset: 0,
        }),
      enabled: !!orgId && !!appId && !!conn.app_branch_id,
    })),
  })

  const runsWithIds = useMemo(() =>
    connections.map((conn, idx) => {
      const runs = runQueries[idx]?.data?.data ?? []
      const run = runs[0]
      const branchRun = run?.app_branch_runs?.at(0)
      return {
        conn,
        run,
        branchRun,
        branchRunId: branchRun?.id,
        branchId: conn.app_branch_id ?? '',
      }
    }),
    [connections, runQueries]
  )

  const buildQueries = useQueries({
    queries: runsWithIds.map(({ branchId, branchRunId }) => ({
      queryKey: ['branch-run-builds', orgId, appId, branchId, branchRunId],
      queryFn: () =>
        getBranchRunBuilds({
          orgId: orgId!,
          appId: appId!,
          branchId,
          runId: branchRunId!,
        }),
      enabled: !!orgId && !!appId && !!branchId && !!branchRunId,
    })),
  })

  const installGroupQueries = useQueries({
    queries: runsWithIds.map(({ branchId, branchRunId }) => ({
      queryKey: ['branch-run-install-groups', orgId, appId, branchId, branchRunId],
      queryFn: () =>
        getBranchRunInstallGroups({
          orgId: orgId!,
          appId: appId!,
          branchId,
          runId: branchRunId!,
        }),
      enabled: !!orgId && !!appId && !!branchId && !!branchRunId,
    })),
  })

  const installId = install?.id ?? ''
  const { data: configVersions } = useQuery({
    queryKey: ['install-app-config-versions', orgId, installId],
    queryFn: () => getInstallAppConfigVersions({ installId: installId!, orgId: orgId! }),
    enabled: !!orgId && !!installId,
  })

  const branches: IBranchEntry[] = useMemo(
    () =>
      runsWithIds.map(({ conn, run, branchRun }, idx) => {
        const configStep = run?.steps?.find(
          (s: any) => s.name?.toLowerCase().includes('config') && !s.name?.toLowerCase().includes('diff')
        )
        const appConfigId = configStep?.status?.metadata?.app_config_id as string | undefined

        const branchVersions = (configVersions ?? []).filter(
          (v) => !!v.app_branch_run_id
        )

        return {
          branchId: conn.app_branch_id ?? '',
          branchName: conn.app_branch?.name ?? conn.app_branch_id ?? 'Unknown',
          active: conn.active ?? false,
          activatedAt: conn.activated_at,
          latestRun: run,
          branchRun,
          builds: buildQueries[idx]?.data ?? [],
          installUpdates: installGroupQueries[idx]?.data ?? [],
          appConfigId,
          configVersions: branchVersions,
        }
      }),
    [runsWithIds, buildQueries, installGroupQueries, configVersions]
  )

  return <InstallBranches branches={branches} orgId={orgId} appId={appId} installId={installId} />
}
