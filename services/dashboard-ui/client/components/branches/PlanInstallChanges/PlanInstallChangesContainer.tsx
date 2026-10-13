import { useMemo } from 'react'
import { useParams } from 'react-router'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useRunTemplateBuilds } from '@/components/branches/BranchOverview/use-run-template-builds'
import { useBranchConfigSections } from '@/components/branches/BranchRunChangesSummary'
import { usePlanGroupInstalls } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep'
import type { TInstallWorkflowStep } from '@/types'
import { mergeBranchSections } from './merge-branch-sections'
import {
  PlanInstallChanges,
  type TPlanInstallFacts,
} from './PlanInstallChanges'

interface IPlanInstallChangesContainer {
  step: TInstallWorkflowStep
  installFacts?: TPlanInstallFacts
  branchRunId?: string
  workflowSteps?: TInstallWorkflowStep[]
}

export const PlanInstallChangesContainer = ({
  step,
  installFacts,
  branchRunId,
  workflowSteps = [],
}: IPlanInstallChangesContainer) => {
  const { org } = useOrg()
  const { app, labelColors } = useApp()
  const { branchId } = useParams()
  const orgId = org?.id
  const appId = app?.id

  const { installs, isLoading } = usePlanGroupInstalls(
    step,
    (step.status?.metadata ?? {}) as Record<string, any>
  )

  const { metaBuilds, templateBuilds } = useRunTemplateBuilds({
    orgId,
    appId,
    branchId,
    branchRunId,
    workflowSteps,
  })

  const configSections = useBranchConfigSections({
    branchId,
    appBranchRunId: branchRunId,
    builds: metaBuilds,
    enabled: !!branchRunId,
  })

  const mergedInstalls = useMemo(
    () =>
      installs.map((install) =>
        install.alreadyApplied
          ? install
          : {
              ...install,
              sections: mergeBranchSections(
                install.sections,
                configSections.sections
              ),
            }
      ),
    [installs, configSections.sections]
  )

  return (
    <PlanInstallChanges
      installs={mergedInstalls}
      builds={templateBuilds}
      installFacts={installFacts}
      labelColors={labelColors}
      isLoading={isLoading}
    />
  )
}
