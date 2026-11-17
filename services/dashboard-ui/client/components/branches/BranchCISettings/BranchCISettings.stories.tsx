export default {
  title: 'Features / Branches / Branch CI settings',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { BranchCISettingsCard, BranchCISettingsModal } from './BranchCISettings'

export const Card = () => (
  <BranchCISettingsCard
    ignoreChangesRegex="^(docs/|README\.md)"
    sendStatusesOnIgnore
    pathFilterMode="paths"
    pathPatterns={'apps/acme/**\nshared/helm/**'}
    onEdit={() => {}}
  />
)

export const Disabled = () => (
  <BranchCISettingsCard
    ignoreChangesRegex=""
    sendStatusesOnIgnore={false}
    pathFilterMode="paths"
    pathPatterns=""
    onEdit={() => {}}
  />
)

export const IgnoredPaths = () => (
  <BranchCISettingsCard
    ignoreChangesRegex=""
    sendStatusesOnIgnore
    pathFilterMode="paths_ignore"
    pathPatterns={'docs/**\n**/*.md'}
    onEdit={() => {}}
  />
)

export const Editor = () => (
  <ModalStory>
    <BranchCISettingsModal
      ignoreChangesRegex="^docs/"
      sendStatusesOnIgnore
      pathFilterMode="paths"
      pathPatterns={'apps/acme/**\n!apps/acme/docs/**'}
      isPending={false}
      error={null}
      onSubmit={() => {}}
      onCancel={() => {}}
    />
  </ModalStory>
)
