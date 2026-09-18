export default {
  title: 'Branches/BranchDetailActions',
}

import { BranchDetailActions } from './BranchDetailActions'

const noop = () => {}

export const Default = () => (
  <BranchDetailActions
    isTriggerPending={false}
    onTriggerRun={noop}
    onTriggerPreviewModal={noop}
  />
)

export const TriggerPending = () => (
  <BranchDetailActions
    isTriggerPending={true}
    onTriggerRun={noop}
    onTriggerPreviewModal={noop}
  />
)

export const WithNudge = () => (
  <BranchDetailActions
    isTriggerPending={false}
    showTriggerNudge
    onTriggerRun={noop}
    onTriggerPreviewModal={noop}
  />
)
