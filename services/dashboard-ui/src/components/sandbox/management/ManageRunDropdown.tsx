'use client'

import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { RunnerJobPlanButton } from '@/components/runners/RunnerJobPlan'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import { DriftScanSandboxButton } from './DriftScanSandbox'
import { ReprovisionSandboxButton } from './ReprovisionSandbox'
import { SandboxRunOutputsButton } from './SandboxRunOutputsModal'
import { useSandboxRun } from '@/hooks/use-sandbox-run'
import type { TWorkflow } from '@/types'

interface ManageRunDropdownProps extends Omit<IDropdown, 'id' | 'buttonText' | 'children'> {
  workflow?: TWorkflow
}

export const ManageRunDropdown = ({
  workflow,
  alignment = 'right',
  ...props
}: ManageRunDropdownProps) => {
  const { sandboxRun } = useSandboxRun()
  
  const hasRunnerJobs = sandboxRun?.runner_jobs?.length > 0
  const shouldShowCancel = sandboxRun?.runner_jobs &&
    sandboxRun?.status_v2?.status !== 'active' &&
    sandboxRun?.status_v2?.status !== 'error' &&
    workflow &&
    !workflow?.finished

  return (
    <Dropdown
      id={`sandbox-run-${sandboxRun.id}-mgmt`}
      variant="primary"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage
        </>
      }
      alignment={alignment}
      {...props}
    >
      <Menu>
        <Text>Settings</Text>
        {hasRunnerJobs && sandboxRun?.runner_jobs?.[0] ? (
          <RunnerJobPlanButton
            isMenuButton
            buttonText="Sync plan"
            runnerJobId={sandboxRun.runner_jobs[0].id}
          />
        ) : null}
        {hasRunnerJobs && sandboxRun?.runner_jobs?.[1] ? (
          <RunnerJobPlanButton
            isMenuButton
            buttonText="Deploy plan"
            runnerJobId={sandboxRun.runner_jobs[1].id}
          />
        ) : null}
        <SandboxRunOutputsButton
          isMenuButton
          sandboxRun={sandboxRun}
        />
        <hr />
        <Text>Controls</Text>
        <DriftScanSandboxButton isMenuButton />
        <ReprovisionSandboxButton isMenuButton />

        {shouldShowCancel ? (
          <CancelWorkflowButton
            className="!text-red-600 dark:!text-red-400"
            isMenuButton
            workflow={workflow}
          />
        ) : null}
      </Menu>
    </Dropdown>
  )
}