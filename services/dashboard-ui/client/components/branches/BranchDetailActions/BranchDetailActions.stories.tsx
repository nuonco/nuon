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
    manageInstallsButton={<MockManageButton />}
    hasConfig={true}
    isTriggerPending={false}
    isCheckPending={false}
    onTriggerRun={noop}
    onCheckForUpdates={noop}
  />
)

export const NoConfig = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    manageInstallsButton={<MockManageButton />}
    hasConfig={false}
    isTriggerPending={false}
    isCheckPending={false}
    onTriggerRun={noop}
    onCheckForUpdates={noop}
  />
)

export const TriggerPending = () => (
  <BranchDetailActions
    editButton={<MockEditButton />}
    manageInstallsButton={<MockManageButton />}
    hasConfig={true}
    isTriggerPending={true}
    isCheckPending={false}
    onTriggerRun={noop}
    onCheckForUpdates={noop}
  />
)
