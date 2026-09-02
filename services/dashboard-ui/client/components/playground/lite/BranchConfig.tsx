import { useState } from 'react'
import { useParams } from 'react-router'
import { Block } from './Block'
import { CompareDrawer } from './CompareDrawer'
import { Drawer } from './Drawer'
import { InstallReadme } from './InstallReadme'
import { NodeDrawer } from './NodeDrawer'
import { ConfigSection } from './ConfigSection'
import { Panel } from './Panel'
import { nodePath } from './BranchOverview'
import { branchBase } from './nav'
import { branchConfigItems, runRows, type IBranchConfigItem } from './fixtures'
import { rowHoverClass } from './utils'

const sections: {
  title: string
  kind: IBranchConfigItem['kind']
  filters?: string[]
}[] = [
  { title: 'Stack', kind: 'stack' },
  { title: 'Roles', kind: 'role', filters: ['Trust'] },
  { title: 'Sandbox', kind: 'sandbox' },
  { title: 'Components', kind: 'component', filters: ['Type', 'Status'] },
  { title: 'Actions', kind: 'action', filters: ['Trigger', 'Status'] },
  { title: 'Runbooks', kind: 'runbook', filters: ['Trigger', 'Status'] },
  { title: 'Policies', kind: 'policy', filters: ['Applies to', 'Severity'] },
]

export const BranchConfig = () => {
  const { appId = '', branchId = '' } = useParams()
  const base = branchBase(appId, branchId)
  const [selected, setSelected] = useState<IBranchConfigItem | undefined>()
  const [isHistoryOpen, setIsHistoryOpen] = useState(false)
  const [isCompareOpen, setIsCompareOpen] = useState(false)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between gap-4">
        <Block
          className="h-[32px] cursor-pointer"
          title="Config version"
          icon="ClockCounterClockwiseIcon"
          text="Current — run-08"
          onClick={() => setIsHistoryOpen(true)}
        />
        <Block
          className="h-[32px] cursor-pointer opacity-60"
          title="Compare versions"
          icon="SplitHorizontalIcon"
          text="Compare"
          onClick={() => setIsCompareOpen(true)}
        />
      </div>

      {sections.map((section) => {
        const items = branchConfigItems.filter(
          (item) => item.kind === section.kind
        )

        if (items.length === 0) return null

        return (
          <ConfigSection
            key={section.title}
            title={section.title}
            items={items}
            filters={section.filters}
            onSelect={setSelected}
          />
        )
      })}

      <Panel title="Readme" action="View source">
        <InstallReadme />
      </Panel>

      {selected && (
        <NodeDrawer
          title={selected.label}
          path={nodePath(base, {
            ...selected,
            x: 0,
            y: 0,
            width: 0,
            height: 0,
          })}
          onClose={() => setSelected(undefined)}
        />
      )}

      {isCompareOpen && (
        <CompareDrawer onClose={() => setIsCompareOpen(false)} />
      )}

      {isHistoryOpen && (
        <Drawer title="Config history" onClose={() => setIsHistoryOpen(false)}>
          <div className="flex flex-col gap-2">
            {runRows.map((width, i) => (
              <div
                key={i}
                className={`flex items-center gap-4 ${rowHoverClass}`}
                title={`run-${i + 1}`}
              >
                <Block className="h-[12px] w-[12px] flex-none rounded-full" />
                <div className="flex min-w-0 flex-1 flex-col gap-1.5">
                  <Block className="h-[10px]" style={{ width }} />
                  <Block className="h-[8px] w-[110px] opacity-50" />
                </div>
              </div>
            ))}
          </div>
        </Drawer>
      )}
    </div>
  )
}
