export default {
  title: 'Installs/BranchStep',
}

import { useState } from 'react'
import { BranchStep } from './BranchStep'
import type { TAppBranch } from '@/types'

const noop = () => {}

const mockBranches: TAppBranch[] = [
  {
    id: 'branch-1',
    name: 'main',
    configs: [
      {
        id: 'cfg-1',
        install_groups: [
          { id: 'g1', name: 'Production', label_selector: { match_labels: { env: 'production' } } },
          { id: 'g2', name: 'Staging', label_selector: { match_labels: { env: 'staging' } } },
        ],
      },
    ],
  } as unknown as TAppBranch,
  {
    id: 'branch-2',
    name: 'develop',
    configs: [
      {
        id: 'cfg-2',
        install_groups: [
          { id: 'g3', name: 'Dev', label_selector: { match_labels: { env: 'dev' } } },
        ],
      },
    ],
  } as unknown as TAppBranch,
  {
    id: 'branch-3',
    name: 'feature/new-ui',
    configs: [{ id: 'cfg-3', install_groups: [] }],
  } as unknown as TAppBranch,
]

export const Default = () => {
  const [selected, setSelected] = useState<TAppBranch | null>(null)
  return (
    <div className="p-6 max-w-lg">
      <BranchStep
        branches={mockBranches}
        selected={selected}
        onSelect={setSelected}
        onSkip={noop}
      />
    </div>
  )
}

export const WithSelection = () => {
  const [selected, setSelected] = useState<TAppBranch | null>(mockBranches[0])
  return (
    <div className="p-6 max-w-lg">
      <BranchStep
        branches={mockBranches}
        selected={selected}
        onSelect={setSelected}
        onSkip={noop}
      />
    </div>
  )
}

export const Empty = () => (
  <div className="p-6 max-w-lg">
    <BranchStep
      branches={[]}
      selected={null}
      onSelect={noop}
      onSkip={noop}
    />
  </div>
)
