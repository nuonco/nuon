'use client'

import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { DeployComponentButton } from '@/components/install-components/management/DeployComponent'
import { DriftScanComponentButton } from '@/components/install-components/management/DriftScanComponent'
import { RunnerJobPlanButton } from '@/components/runners/RunnerJobPlan'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import { useDeploy } from '@/hooks/use-deploy'
import type { TComponent, TWorkflow } from '@/types'

export const ManagementDropdown = ({
  component,
  currentBuildId,
  workflow,
}: {
  component: TComponent
  currentBuildId?: string
  workflow: TWorkflow
}) => {
  const { deploy } = useDeploy()
  return (
    <Dropdown
      id={`component-${component.id}-mgmt`}
      variant="primary"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Manage
        </>
      }
      alignment="right"
    >
      <Menu>
        <Text>Settings</Text>
        {deploy?.runner_jobs?.length && deploy?.runner_jobs?.[0] ? (
          <RunnerJobPlanButton
            isMenuButton
            buttonText="Sync plan"
            runnerJobId={deploy?.runner_jobs?.[0]?.id}
          />
        ) : null}
        {deploy?.runner_jobs?.length && deploy?.runner_jobs?.[1] ? (
          <RunnerJobPlanButton
            isMenuButton
            buttonText="Deploy plan"
            runnerJobId={deploy?.runner_jobs?.[1]?.id}
          />
        ) : null}
        <hr />
        <Text>Controls</Text>
        <DriftScanComponentButton
          component={component}
          currentBuildId={currentBuildId}
          isMenuButton
        />
        <DeployComponentButton
          component={component}
          currentBuildId={currentBuildId}
          isMenuButton
        />

        {deploy?.runner_jobs &&
        deploy?.status_v2?.status !== 'active' &&
        deploy?.status_v2?.status !== 'error' &&
        workflow &&
        !workflow?.finished ? (
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
