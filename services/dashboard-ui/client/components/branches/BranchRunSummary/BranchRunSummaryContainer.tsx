import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstallsById } from '@/hooks/use-installs-by-id'
import { useOrg } from '@/hooks/use-org'
import { getBranchRunBuilds, getBranchRunInstallGroups, getBranchInstallGroupRuns, getSandboxBuilds } from '@/lib'
import { BranchRunSummary } from './BranchRunSummary'
import type { TAppBranchRun } from '@/types'

interface IBranchRunSummaryContainer {
  branchRun?: TAppBranchRun
  appId: string
  branchId: string
  branchRunId?: string
  runStatus: string
}

const TERMINAL = new Set(['success', 'failed', 'cancelled'])

export const BranchRunSummaryContainer = ({
  branchRun,
  appId,
  branchId,
  branchRunId,
  runStatus,
}: IBranchRunSummaryContainer) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''
  const installsById = useInstallsById(appId)
  const isTerminal = TERMINAL.has(runStatus)

  const { data: builds } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-run-builds', orgId, appId, branchId, branchRunId],
    queryFn: () => getBranchRunBuilds({ orgId: orgId!, appId, branchId, runId: branchRunId! }),
    enabled: !!orgId && !!branchRunId,
    refetchInterval: isTerminal ? false : 5000,
  })

  const { data: installUpdates } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-run-install-groups', orgId, appId, branchId, branchRunId],
    queryFn: () => getBranchRunInstallGroups({ orgId: orgId!, appId, branchId, runId: branchRunId! }),
    enabled: !!orgId && !!branchRunId,
    refetchInterval: isTerminal ? false : 5000,
  })

  const { data: installGroupRuns } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['branch-install-group-runs', orgId, appId, branchId, branchRunId],
    queryFn: () => getBranchInstallGroupRuns({ orgId: orgId!, appId, branchId, runId: branchRunId! }),
    enabled: !!orgId && !!branchRunId,
    refetchInterval: isTerminal ? false : 5000,
  })

  const { data: sandboxBuilds } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['sandbox-builds', orgId, appId],
    queryFn: () => getSandboxBuilds({ orgId: orgId!, appId, limit: 50 }),
    enabled: !!orgId && !!appId && !!branchRun?.app_config_id,
    refetchInterval: isTerminal ? false : 5000,
  })

  const sandboxBuild =
    sandboxBuilds?.data?.find(
      (b) => b.app_config_id === branchRun?.app_config_id
    ) ?? null

  return (
    <BranchRunSummary
      branchRun={branchRun}
      builds={builds ?? []}
      sandboxBuild={sandboxBuild}
      installUpdates={installUpdates ?? []}
      installGroupRuns={installGroupRuns ?? []}
      installsById={installsById}
      orgId={orgId}
      appId={appId}
      branchId={branchId}
      runStatus={runStatus}
    />
  )
}
