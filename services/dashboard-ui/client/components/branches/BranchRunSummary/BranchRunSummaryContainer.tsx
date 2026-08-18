import { useQuery } from '@tanstack/react-query'
import { useInstallsById } from '@/hooks/use-installs-by-id'
import { useOrg } from '@/hooks/use-org'
import { getBranchRunBuilds, getBranchRunInstallGroups, getBranchInstallGroupRuns } from '@/lib'
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
    queryKey: ['branch-run-builds', orgId, appId, branchId, branchRunId],
    queryFn: () => getBranchRunBuilds({ orgId: orgId!, appId, branchId, runId: branchRunId! }),
    enabled: !!orgId && !!branchRunId,
    refetchInterval: isTerminal ? false : 5000,
  })

  const { data: installUpdates } = useQuery({
    queryKey: ['branch-run-install-groups', orgId, appId, branchId, branchRunId],
    queryFn: () => getBranchRunInstallGroups({ orgId: orgId!, appId, branchId, runId: branchRunId! }),
    enabled: !!orgId && !!branchRunId,
    refetchInterval: isTerminal ? false : 5000,
  })

  const { data: installGroupRuns } = useQuery({
    queryKey: ['branch-install-group-runs', orgId, appId, branchId, branchRunId],
    queryFn: () => getBranchInstallGroupRuns({ orgId: orgId!, appId, branchId, runId: branchRunId! }),
    enabled: !!orgId && !!branchRunId,
    refetchInterval: isTerminal ? false : 5000,
  })

  return (
    <BranchRunSummary
      branchRun={branchRun}
      builds={builds ?? []}
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
