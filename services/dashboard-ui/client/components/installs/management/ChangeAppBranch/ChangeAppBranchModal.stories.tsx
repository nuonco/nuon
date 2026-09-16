export default {
  title: 'Installs/ChangeAppBranchModal',
}

import { useState } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import { ChangeAppBranchModal } from './ChangeAppBranchModal'
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
          { id: 'g1', name: 'Prod', label_selector: { match_labels: { env: 'production' } } },
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
          { id: 'g2', name: 'Staging', label_selector: { match_labels: { env: 'staging' } } },
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

export const Default = () => {
  const [target, setTarget] = useState<TAppBranch | null>(null)
  return (
    <ModalStory>
      <ChangeAppBranchModal
        install={mockInstall}
        targetBranch={target}
        branches={mockBranches}
        isPending={false}
        onSelectBranch={setTarget}
        onConfirm={noop}
        onClose={noop}
      />
    </ModalStory>
  )
}

export const WithTargetSelected = () => {
  const [target, setTarget] = useState<TAppBranch | null>(mockBranches[1])
  return (
    <ModalStory>
      <ChangeAppBranchModal
        install={mockInstall}
        targetBranch={target}
        branches={mockBranches}
        isPending={false}
        onSelectBranch={setTarget}
        onConfirm={noop}
        onClose={noop}
      />
    </ModalStory>
  )
}

export const NoMatchingGroupWarning = () => {
  const [target, setTarget] = useState<TAppBranch | null>(mockBranches[2])
  return (
    <ModalStory>
      <ChangeAppBranchModal
        install={mockInstall}
        targetBranch={target}
        branches={mockBranches}
        isPending={false}
        onSelectBranch={setTarget}
        onConfirm={noop}
        onClose={noop}
      />
    </ModalStory>
  )
}

export const Pending = () => {
  const [target] = useState<TAppBranch | null>(mockBranches[1])
  return (
    <ModalStory>
      <ChangeAppBranchModal
        install={mockInstall}
        targetBranch={target}
        branches={mockBranches}
        isPending
        onSelectBranch={noop}
        onConfirm={noop}
        onClose={noop}
      />
    </ModalStory>
  )
}

export const NoBranchInstall = () => {
  const installNoBranch: TInstall = {
    ...mockInstall,
    app_branch_id: undefined,
    app_branch: undefined,
  } as unknown as TInstall
  return (
    <ModalStory>
      <ChangeAppBranchModal
        install={installNoBranch}
        targetBranch={null}
        branches={mockBranches}
        isPending={false}
        onSelectBranch={noop}
        onConfirm={noop}
        onClose={noop}
      />
    </ModalStory>
  )
}
