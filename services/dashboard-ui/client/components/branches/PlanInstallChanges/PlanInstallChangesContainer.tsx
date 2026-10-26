import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getBranchWorkflowRun } from '@/lib'
import { useRunTemplateBuilds } from '@/components/branches/BranchOverview/use-run-template-builds'
import { usePlanGroupInstalls } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep'
import type { TInstallWorkflowStep } from '@/types'
import {
  PlanInstallChanges,
  type TPlanInstallFacts,
} from './PlanInstallChanges'

interface IPlanInstallChangesContainer {
  step: TInstallWorkflowStep
  installFacts?: TPlanInstallFacts
}

export const PlanInstallChangesContainer = ({
  step,
  installFacts,
}: IPlanInstallChangesContainer) => {
  const { org } = useOrg()
  const { app, labelColors } = useApp()
  const { branchId } = useParams()
  const orgId = org?.id
  const appId = app?.id
  const workflowId = step.install_workflow_id
  const { installs, isLoading } = usePlanGroupInstalls(
    step,
    (step.status?.metadata ?? {}) as Record<string, any>
  )
  const runEnabled = !!orgId && !!appId && !!branchId && !!workflowId
  const {
    data: run,
    isPending: runPending,
    isPlaceholderData: runPlaceholder,
    isError: runError,
  } = useQuery({
    queryKey: ['branch-run', orgId, appId, branchId, workflowId],
    queryFn: () =>
      getBranchWorkflowRun({
        orgId: orgId!,
        appId: appId!,
        branchId: branchId!,
        runId: workflowId!,
      }),
    enabled: runEnabled,
    placeholderData: keepPreviousData,
  })
  const branchRunId = run?.app_branch_runs?.at(0)?.id
  const { templateBuilds, buildsPending } = useRunTemplateBuilds({
    orgId,
    appId,
    branchId,
    branchRunId,
    workflowSteps: run?.steps ?? [],
  })
  const waitingForRun = runEnabled && !runError && (runPending || runPlaceholder)

  return (
    <PlanInstallChanges
      installs={installs}
      builds={templateBuilds}
      installFacts={installFacts}
      labelColors={labelColors}
      isLoading={isLoading || waitingForRun || (!!branchRunId && buildsPending)}
    />
  )
}
