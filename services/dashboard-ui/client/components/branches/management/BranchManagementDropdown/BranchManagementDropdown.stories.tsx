export default {
  title: 'Branches/BranchManagementDropdown',
}

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { BranchManagementDropdown } from './BranchManagementDropdown'

const noop = () => {}

const EditButton = () => (
  <Button isMenuButton onClick={noop}>
    Edit branch
    <Icon variant="PencilSimpleLineIcon" size={16} />
  </Button>
)

const DeploymentPlanButton = () => (
  <Button isMenuButton onClick={noop}>
    Deployment plan
    <Icon variant="SlidersHorizontalIcon" size={16} />
  </Button>
)

const Wrap = ({ children }: { children: React.ReactNode }) => (
  <div className="flex justify-end p-8">{children}</div>
)

export const Default = () => (
  <Wrap>
    <BranchManagementDropdown
      dropdownId="br-001"
      detailHref="/org-1/apps/app-1/branches/br-001"
      editButton={<EditButton />}
      deploymentPlanButton={<DeploymentPlanButton />}
      hasConfig
      isTriggerPending={false}
      onTriggerRun={noop}
    />
  </Wrap>
)

export const NoConfig = () => (
  <Wrap>
    <BranchManagementDropdown
      dropdownId="br-002"
      detailHref="/org-1/apps/app-1/branches/br-002"
      editButton={<EditButton />}
      deploymentPlanButton={<DeploymentPlanButton />}
      hasConfig={false}
      isTriggerPending={false}
      onTriggerRun={noop}
    />
  </Wrap>
)

export const TriggerPending = () => (
  <Wrap>
    <BranchManagementDropdown
      dropdownId="br-003"
      detailHref="/org-1/apps/app-1/branches/br-003"
      editButton={<EditButton />}
      deploymentPlanButton={<DeploymentPlanButton />}
      hasConfig
      isTriggerPending
      onTriggerRun={noop}
    />
  </Wrap>
)
