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
    isTriggerPending={false}
    onTriggerRun={noop}
    onTriggerPreviewModal={noop}
  />
)

export const TriggerPending = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    deploymentPlanButton={<MockDeploymentPlanButton />}
    deleteButton={<MockDeleteButton />}
    isTriggerPending={true}
    onTriggerRun={noop}
    onTriggerPreviewModal={noop}
  />
)

export const WithNudge = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    deploymentPlanButton={<MockDeploymentPlanButton />}
    deleteButton={<MockDeleteButton />}
    isTriggerPending={false}
    showTriggerNudge
    onTriggerRun={noop}
    onTriggerPreviewModal={noop}
  />
)
