import { useState } from 'react'
import { useParams } from 'react-router'
import { Block } from './Block'
import { BranchTile } from './BranchTile'
import { Drawer } from './Drawer'
import { InstallGraph, type IGraphNode } from './InstallGraph'
import { InstallReadme } from './InstallReadme'
import { LastUpdateTile } from './LastUpdateTile'
import { LinkBlock } from './LinkBlock'
import { Panel } from './Panel'
import { StatTile } from './StatTile'
import { labelWidth } from './utils'

const nodePath = (installId: string, node: IGraphNode) => {
  switch (node.kind) {
    case 'install':
      return `/installs/${installId}/details`
    case 'stack':
      return `/installs/${installId}/stack`
    case 'access':
      return `/installs/${installId}/access`
    case 'sandbox':
      return `/installs/${installId}/sandbox`
    case 'runner':
      return `/installs/${installId}/runner`
    case 'component':
      return `/installs/${installId}/components/${node.id}`
    default:
      return undefined
  }
}

export const InstallOverview = () => {
  const { installId = '' } = useParams()
  const [selected, setSelected] = useState<IGraphNode | undefined>()
  const path = selected ? nodePath(installId, selected) : undefined

  return (
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatTile label="Health" valueWidth={64} />
        <StatTile label="Drift" valueWidth={72} />
        <BranchTile />
        <LastUpdateTile />
      </div>

      <Panel title="Install graph" action="Refresh">
        <InstallGraph selectedId={selected?.id} onSelect={setSelected} />
      </Panel>

      <Panel title="Readme" action="Edit">
        <InstallReadme />
      </Panel>

      {selected && (
        <Drawer title={selected.label} onClose={() => setSelected(undefined)}>
          <div className="flex flex-col gap-3">
            <Block className="h-[12px] w-[70%]" />
            <Block className="h-[8px] w-[45%] opacity-50" />
          </div>

          <div className="flex flex-col gap-4">
            {['Status', 'Type', 'Updated', 'Version'].map((label) => (
              <div key={label} className="flex items-center justify-between">
                <Block
                  className="h-[8px] opacity-50"
                  style={{ width: labelWidth(label) }}
                  title={label}
                  text={label}
                />
                <Block className="h-[10px] w-[96px]" />
              </div>
            ))}
          </div>

          {path && (
            <LinkBlock
              path={path}
              label={`View ${selected.label.toLowerCase()}`}
              className="h-[32px] w-full"
              style={{ width: '100%' }}
            />
          )}
        </Drawer>
      )}
    </div>
  )
}
