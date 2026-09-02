export default {
  title: 'Branches/BranchCISettings',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { BranchCISettingsCard, BranchCISettingsModal } from './BranchCISettings'

export const Card = () => (
  <BranchCISettingsCard
    ignoreChangesRegex="^(docs/|README\.md)"
    sendStatusesOnIgnore
    onEdit={() => {}}
  />
)

export const Disabled = () => (
  <BranchCISettingsCard
    ignoreChangesRegex=""
    sendStatusesOnIgnore={false}
    onEdit={() => {}}
  />
)

export const Editor = () => (
  <ModalStory>
    <BranchCISettingsModal
      ignoreChangesRegex="^docs/"
      sendStatusesOnIgnore
      isPending={false}
      error={null}
      onSubmit={() => {}}
      onCancel={() => {}}
    />
  </ModalStory>
)
