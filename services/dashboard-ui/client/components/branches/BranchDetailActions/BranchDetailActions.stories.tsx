export default {
  title: 'Branches/BranchDetailActions',
}

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { BranchDetailActions } from './BranchDetailActions'

const noop = () => {}

const MockEditButton = () => (
  <Button isMenuButton>
    Edit branch
    <Icon variant="PencilSimpleLineIcon" size={16} />
  </Button>
)

const MockDeploymentPlanButton = () => (
  <Button isMenuButton>
    Deployment plan
    <Icon variant="SlidersHorizontalIcon" size={16} />
  </Button>
)

const MockDeleteButton = () => (
  <Button isMenuButton variant="danger">
    Delete branch
    <Icon variant="TrashIcon" size={16} />
  </Button>
)

export const Default = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    deploymentPlanButton={<MockDeploymentPlanButton />}
    deleteButton={<MockDeleteButton />}
    hasConfig={true}
    isTriggerPending={false}
    onTriggerRun={noop}
    onTriggerPreview={noop}
  />
)

export const NoConfig = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    deploymentPlanButton={<MockDeploymentPlanButton />}
    deleteButton={<MockDeleteButton />}
    hasConfig={false}
    isTriggerPending={false}
    onTriggerRun={noop}
    onTriggerPreview={noop}
  />
)

export const TriggerPending = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    deploymentPlanButton={<MockDeploymentPlanButton />}
    deleteButton={<MockDeleteButton />}
    hasConfig={true}
    isTriggerPending={true}
    onTriggerRun={noop}
    onTriggerPreview={noop}
  />
)
