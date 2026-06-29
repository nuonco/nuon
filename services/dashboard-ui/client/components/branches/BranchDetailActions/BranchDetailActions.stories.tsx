export default {
  title: 'Branches/BranchDetailActions',
}

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { BranchDetailActions } from './BranchDetailActions'

const noop = () => {}

const MockEditButton = () => (
  <Button variant="secondary">
    <Icon variant="PencilSimpleLineIcon" size={16} />
    Edit branch
  </Button>
)

const MockManageButton = () => (
  <Button variant="secondary">
    <Icon variant="SlidersHorizontalIcon" size={16} />
    Manage installs
  </Button>
)

export const Default = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    deleteButton={null}
    manageInstallsButton={<MockManageButton />}
    hasConfig={true}
    isTriggerPending={false}
    onTriggerRun={noop}
    onTriggerPreview={noop}
  />
)

export const NoConfig = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    deleteButton={null}
    manageInstallsButton={<MockManageButton />}
    hasConfig={false}
    isTriggerPending={false}
    onTriggerRun={noop}
    onTriggerPreview={noop}
  />
)

export const TriggerPending = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    deleteButton={null}
    manageInstallsButton={<MockManageButton />}
    hasConfig={true}
    isTriggerPending={true}
    onTriggerRun={noop}
    onTriggerPreview={noop}
  />
)
