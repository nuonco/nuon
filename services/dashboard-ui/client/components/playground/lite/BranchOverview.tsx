import { useState } from 'react'
import { useParams } from 'react-router'
import { BranchGraph, type IBranchNode } from './BranchGraph'
import { BranchTile } from './BranchTile'
import { LastUpdateTile } from './LastUpdateTile'
import { NodeDrawer } from './NodeDrawer'
import { Panel } from './Panel'
import { StatTile } from './StatTile'
import { branchBase } from './nav'

export const nodePath = (base: string, node: IBranchNode) => {
  switch (node.kind) {
    case 'source':
      return `${base}/details`
    case 'component':
      return `${base}/components/${node.id}`
    case 'stack':
      return `${base}/stack`
    case 'sandbox':
      return `${base}/sandbox`
    case 'action':
      return `${base}/actions/${node.id}`
    case 'runbook':
      return `${base}/runbooks/${node.id}`
    case 'role':
      return `${base}/roles/${node.id}`
    case 'policy':
      return `${base}/policies/${node.id}`
    case 'build':
      return `${base}/builds`
    case 'group':
      return `${base}/groups/${node.id}`
    case 'install':
      return `/installs/${node.id}`
    default:
      return undefined
  }
}

export const BranchOverview = () => {
  const { appId = '', branchId = '' } = useParams()
  const base = branchBase(appId, branchId)
  const [selected, setSelected] = useState<IBranchNode | undefined>()
  const path = selected ? nodePath(base, selected) : undefined

  return (
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-3 gap-4">
        <BranchTile />
        <LastUpdateTile />
        <StatTile label="Installs" valueWidth={56} />
      </div>

      <Panel title="Branch pipeline" action="Refresh">
        <BranchGraph selectedId={selected?.id} onSelect={setSelected} />
      </Panel>

      {selected && (
        <NodeDrawer
          title={selected.label}
          path={path}
          onClose={() => setSelected(undefined)}
        />
      )}
    </div>
  )
}
