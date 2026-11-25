export default {
  title: 'Features / Branches / Branch CI settings',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { BranchCISettingsCard, BranchCISettingsModal } from './BranchCISettings'

export const Card = () => (
  <BranchCISettingsCard
    ignoreChangesRegex="^(docs/|README\.md)"
    sendStatusesOnIgnore
    allowedPaths={'apps/acme/**\nshared/helm/**'}
    deniedPaths=""
    onEdit={() => {}}
  />
)

export const Disabled = () => (
  <BranchCISettingsCard
    ignoreChangesRegex=""
    sendStatusesOnIgnore={false}
    allowedPaths=""
    deniedPaths=""
    onEdit={() => {}}
  />
)

export const DeniedOnly = () => (
  <BranchCISettingsCard
    ignoreChangesRegex=""
    sendStatusesOnIgnore
    allowedPaths=""
    deniedPaths={'docs/**\n**/*.md'}
    onEdit={() => {}}
  />
)

export const Editor = () => (
  <ModalStory>
    <BranchCISettingsModal
      ignoreChangesRegex="^docs/"
      sendStatusesOnIgnore
      allowedPaths="apps/acme/**"
      deniedPaths=""
      isPending={false}
      error={null}
      onSubmit={() => {}}
      onCancel={() => {}}
    />
  </ModalStory>
)
