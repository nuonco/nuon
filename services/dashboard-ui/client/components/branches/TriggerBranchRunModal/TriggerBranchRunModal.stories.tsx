export default {
  title: 'Features / Branches / Trigger branch run modal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { TriggerBranchRunModal } from './TriggerBranchRunModal'

const noop = () => {}

export const Run = () => (
  <ModalStory label="Trigger run">
    <TriggerBranchRunModal
      branchName="production"
      planOnly={false}
      isPending={false}
      onConfirm={noop}
    />
  </ModalStory>
)

export const Preview = () => (
  <ModalStory label="Trigger preview">
    <TriggerBranchRunModal
      branchName="production"
      planOnly
      isPending={false}
      onConfirm={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory label="Triggering">
    <TriggerBranchRunModal
      branchName="production"
      planOnly={false}
      isPending
      onConfirm={noop}
    />
  </ModalStory>
)
