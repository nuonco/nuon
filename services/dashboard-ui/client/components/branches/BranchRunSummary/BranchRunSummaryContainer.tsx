import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getBranchRunBuilds, getBranchRunInstallGroups } from '@/lib'
import { BranchRunSummary } from './BranchRunSummary'
import type { TAppBranchRun } from '@/types'

interface IBranchRunSummaryContainer {
  branchRun?: TAppBranchRun
  appId: string
  branchId: string
  runId: string
  runStatus: string
}

const TERMINAL_STATUSES = ['success', 'failed', 'cancelled']

export const BranchRunSummaryContainer = ({
  branchRun,
  appId,
  branchId,
  runId,
  runStatus,
}: IBranchRunSummaryContainer) => {
  const { org } = useOrg()
  const orgId = org?.id ?? ''
  const isTerminal = TERMINAL_STATUSES.includes(runStatus)

  const { data: builds } = useQuery({
    queryKey: ['branch-run-builds', orgId, appId, branchId, runId],
    queryFn: () => getBranchRunBuilds({ orgId: orgId!, appId, branchId, runId }),
    enabled: !!orgId && isTerminal,
  })

  const { data: installUpdates } = useQuery({
    queryKey: ['branch-run-install-groups', orgId, appId, branchId, runId],
    queryFn: () => getBranchRunInstallGroups({ orgId: orgId!, appId, branchId, runId }),
    enabled: !!orgId && isTerminal,
  })

  if (!isTerminal) return null

  return (
    <BranchRunSummary
      branchRun={branchRun}
      builds={builds ?? []}
      installUpdates={installUpdates ?? []}
      orgId={orgId}
      appId={appId}
      branchId={branchId}
      runStatus={runStatus}
    />
  )
}
