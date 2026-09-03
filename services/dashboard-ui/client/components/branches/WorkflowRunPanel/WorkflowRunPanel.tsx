import { type Ref } from 'react'
import { Button } from '@/components/common/Button'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { WorkflowStepsPipeline } from '@/components/branches/WorkflowStepsPipeline'
import { WorkflowStepDetail } from '@/components/branches/WorkflowStepDetail'
import type { TInstallWorkflowStep } from '@/types'
import { getWorkflowStepTitle } from '@/utils/workflow-utils'

export interface IWorkflowRunPanel extends IPanel {
  steps: TInstallWorkflowStep[]
  selectedStep: TInstallWorkflowStep | null
  activeStep?: TInstallWorkflowStep
  onSelectStep: (step: TInstallWorkflowStep) => void
  onJumpToActive: () => void
  appBranchId?: string
  appBranchRunId?: string
  runTitle: string
  status: string
  isLoading?: boolean
  stepDetailRef?: Ref<HTMLDivElement>
}

export const WorkflowRunPanel = ({
  steps,
  selectedStep,
  activeStep,
  onSelectStep,
  onJumpToActive,
  appBranchId,
  appBranchRunId,
  runTitle,
  status,
  isLoading = false,
  stepDetailRef,
  ...props
}: IWorkflowRunPanel) => {
  return (
    <Panel
      {...props}
      size="3/4"
      heading={
        <div className="flex items-center gap-2.5 min-w-0">
          <Text variant="base" weight="strong" className="truncate">
            {runTitle}
          </Text>
          <Status status={status} variant="badge" className="shrink-0" />
        </div>
      }
    >
      {isLoading ? (
        <Text theme="neutral">Loading workflow run...</Text>
      ) : (
        <>
          <div className="flex flex-col gap-2">
            <div className="flex items-center justify-between gap-3">
              <Text variant="h3" weight="strong">
                Workflow progress
              </Text>
              {activeStep && (
                <Button variant="secondary" onClick={onJumpToActive}>
                  Jump to active step
                </Button>
              )}
            </div>
            <WorkflowStepsPipeline
              steps={steps}
              selectedStepId={selectedStep?.id}
              onSelectStep={onSelectStep}
            />
          </div>

          {selectedStep && (
            <div className="flex flex-col gap-2">
              <div
                ref={stepDetailRef}
                className="flex items-baseline gap-3 scroll-mt-4"
              >
                <Text variant="h3" weight="strong">
                  Step details
                </Text>
                <Text variant="subtext" theme="neutral">
                  {getWorkflowStepTitle(selectedStep)}
                </Text>
              </div>
              <WorkflowStepDetail
                step={selectedStep}
                appBranchId={appBranchId}
                appBranchRunId={appBranchRunId}
                onClose={() => {}}
              />
            </div>
          )}
        </>
      )}
    </Panel>
  )
}
