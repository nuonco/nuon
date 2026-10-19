export default {
  title: 'Features / Installs / Change install group modal',
}

import { useState } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import type { IModal } from '@/components/surfaces/Modal'
import type { TAppBranchInstallGroup, TInstall } from '@/types'
import { ChangeInstallGroupModal } from './ChangeInstallGroupModal'

const noop = () => {}

const install = {
  id: 'inst-1',
  name: 'acme-west',
  app_id: 'app-1',
  app_branch_id: 'branch-main',
  app_branch_group: 'default',
  labels: { env: 'preview' },
} as TInstall

const groups: TAppBranchInstallGroup[] = [
  {
    id: 'group-default',
    name: 'default',
    default: true,
    label_selector: { match_labels: { env: '*' } },
  },
  {
    id: 'group-preview',
    name: 'preview',
    label_selector: { match_labels: { channel: 'preview' } },
  },
  {
    id: 'group-internal',
    name: 'internal',
  },
]

const labeledInstall = {
  ...install,
  app_branch_group: 'staging',
  labels: { env: 'staging', region: 'us-east-1' },
} as TInstall

const labeledGroups: TAppBranchInstallGroup[] = [
  {
    id: 'group-staging',
    name: 'staging',
    label_selector: { match_labels: { env: 'staging' } },
  },
  {
    id: 'group-preview',
    name: 'preview',
    label_selector: { match_labels: { env: 'preview' } },
  },
]

const InteractiveModal = ({
  currentGroup = 'default',
  availableGroups = groups,
  storyInstall = install,
  initialTarget = '',
  isPending = false,
  isLoading = false,
  loadFailed = false,
  ...props
}: IModal & {
  currentGroup?: string
  availableGroups?: TAppBranchInstallGroup[]
  storyInstall?: TInstall
  initialTarget?: string
  isPending?: boolean
  isLoading?: boolean
  loadFailed?: boolean
}) => {
  const [targetGroup, setTargetGroup] = useState(initialTarget)
  return (
    <ChangeInstallGroupModal
      install={storyInstall}
      branchName="main"
      currentGroup={currentGroup}
      groups={availableGroups}
      targetGroup={targetGroup}
      isPending={isPending}
      isLoading={isLoading}
      loadFailed={loadFailed}
      onSelectGroup={setTargetGroup}
      onConfirm={noop}
      {...props}
    />
  )
}

export const Default = () => (
  <ModalStory label="Move acme-west out of default">
    <InteractiveModal />
  </ModalStory>
)

export const PreviewSelected = () => (
  <ModalStory label="Preview already selected">
    <InteractiveModal initialTarget="preview" />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory label="Move in progress">
    <InteractiveModal initialTarget="preview" isPending />
  </ModalStory>
)

export const OnlyGroup = () => (
  <ModalStory label="No other groups">
    <InteractiveModal availableGroups={[groups[0]]} />
  </ModalStory>
)

export const LabeledGroups = () => (
  <ModalStory label="Move between labeled groups">
    <InteractiveModal
      storyInstall={labeledInstall}
      currentGroup="staging"
      availableGroups={labeledGroups}
      initialTarget="preview"
    />
  </ModalStory>
)

export const Unassigned = () => (
  <ModalStory label="Install in no group">
    <InteractiveModal currentGroup="" />
  </ModalStory>
)
