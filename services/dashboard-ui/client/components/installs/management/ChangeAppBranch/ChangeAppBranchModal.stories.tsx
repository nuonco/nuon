export default {
  title: 'Installs/ChangeAppBranchModal',
}

import { useState } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import type { IModal } from '@/components/surfaces/Modal'
import {
  ChangeAppBranchModal,
  type TBranchGroupAssignmentMode,
} from './ChangeAppBranchModal'
import type { TAppBranch, TInstall } from '@/types'

const noop = () => {}

const mockInstall: TInstall = {
  id: 'install-1',
  name: 'acme-prod',
  app_id: 'app-1',
  app_branch_id: 'branch-main',
  app_branch: { id: 'branch-main', name: 'main' } as TAppBranch,
  labels: { env: 'production', region: 'us-east-1' },
} as unknown as TInstall

const mockBranches: TAppBranch[] = [
  {
    id: 'branch-main',
    name: 'main',
    configs: [
      {
        id: 'cfg-1',
        install_groups: [
          {
            id: 'g1',
            name: 'Prod',
            label_selector: { match_labels: { env: 'production' } },
          },
        ],
      },
    ],
  } as unknown as TAppBranch,
  {
    id: 'branch-develop',
    name: 'develop',
    configs: [
      {
        id: 'cfg-2',
        install_groups: [
          {
            id: 'g2',
            name: 'Staging',
            label_selector: { match_labels: { env: 'staging' } },
          },
          {
            id: 'g3',
            name: 'Manually assigned',
          },
          {
            id: 'g4',
            name: 'Other installs',
            default: true,
          },
        ],
      },
    ],
  } as unknown as TAppBranch,
  {
    id: 'branch-feature',
    name: 'feature/new-ia',
    configs: [{ id: 'cfg-3', install_groups: [] }],
  } as unknown as TAppBranch,
]

const InteractiveModal = ({
  initialTarget = null,
  install = mockInstall,
  isPending = false,
  ...props
}: IModal & {
  initialTarget?: TAppBranch | null
  install?: TInstall
  isPending?: boolean
}) => {
  const [target, setTarget] = useState<TAppBranch | null>(initialTarget)
  const [group, setGroup] = useState('')
  const [assignmentMode, setAssignmentMode] =
    useState<TBranchGroupAssignmentMode | null>(null)
  return (
    <ChangeAppBranchModal
      install={install}
      targetBranch={target}
      targetGroup={group}
      assignmentMode={assignmentMode}
      branches={mockBranches}
      isPending={isPending}
      onSelectBranch={(branch) => {
        setTarget(branch)
        setGroup('')
        setAssignmentMode(null)
      }}
      onSelectGroup={(groupName) => {
        setGroup(groupName)
        const selectedGroup = target?.configs
          ?.at(0)
          ?.install_groups?.find((candidate) => candidate.name === groupName)
        setAssignmentMode(selectedGroup?.default ? 'default' : null)
      }}
      onSelectAssignmentMode={setAssignmentMode}
      onConfirm={noop}
      {...props}
    />
  )
}

export const Default = () => (
  <ModalStory>
    <InteractiveModal />
  </ModalStory>
)

export const WithTargetSelected = () => (
  <ModalStory>
    <InteractiveModal initialTarget={mockBranches[1]} />
  </ModalStory>
)

export const NoGroups = () => (
  <ModalStory>
    <InteractiveModal initialTarget={mockBranches[2]} />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <InteractiveModal initialTarget={mockBranches[1]} isPending />
  </ModalStory>
)

export const NoBranchInstall = () => {
  const installNoBranch: TInstall = {
    ...mockInstall,
    app_branch_id: undefined,
    app_branch: undefined,
  } as unknown as TInstall
  return (
    <ModalStory>
      <InteractiveModal install={installNoBranch} />
    </ModalStory>
  )
}
