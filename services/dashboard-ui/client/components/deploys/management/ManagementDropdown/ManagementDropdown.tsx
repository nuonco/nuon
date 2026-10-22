import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { RunnerJobPlanButton } from '@/components/runners/RunnerJobPlan'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import type { TComponent, TDeploy, TWorkflow } from '@/types'

interface IManagementDropdown {
  component: TComponent
  workflow: TWorkflow
  deploy: TDeploy
}

export const ManagementDropdown = ({
  component,
  workflow,
  deploy,
}: IManagementDropdown) => {
  return (
    <Dropdown
      id={`component-${component.id}-mgmt`}
      variant="secondary"
      buttonText={
        <>
          <Icon variant="SlidersHorizontalIcon" /> Deploy controls
        </>
      }
      alignment="right"
    >
      <Menu>
        <Text>Settings</Text>
        {deploy?.runner_jobs?.length && deploy?.runner_jobs?.[1] ? (
          <RunnerJobPlanButton
            isMenuButton
            buttonText="Sync plan"
            runnerJobId={deploy?.runner_jobs?.[1]?.id}
          />
        ) : null}
        {deploy?.runner_jobs?.length && deploy?.runner_jobs?.[0] ? (
          <RunnerJobPlanButton
            isMenuButton
            buttonText="Deploy plan"
            runnerJobId={deploy?.runner_jobs?.[0]?.id}
          />
        ) : null}
        {deploy?.runner_jobs &&
        deploy?.status_v2?.status !== 'active' &&
        deploy?.status_v2?.status !== 'error' &&
        workflow &&
        !workflow?.finished ? (
          <CancelWorkflowButton
            variant="danger"
            isMenuButton
            workflow={workflow}
          />
        ) : null}
      </Menu>
    </Dropdown>
  )
}
