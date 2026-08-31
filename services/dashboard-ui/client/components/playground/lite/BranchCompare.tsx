import { useState } from 'react'
import { useParams } from 'react-router'
import { cn } from '@/utils/classnames'
import { Block } from './Block'
import { Drawer } from './Drawer'
import { Panel } from './Panel'
import { diffEntries, type IDiffEntry } from './fixtures'
import { rowHoverClass } from './utils'

const kindLabel: Record<IDiffEntry['kind'], string> = {
  added: 'Added',
  removed: 'Removed',
  changed: 'Changed',
}

const configLines = ['88%', '64%', '92%', '48%', '76%', '58%']

const DiffColumn = ({ label, muted }: { label: string; muted?: boolean }) => (
  <div className="flex flex-1 flex-col gap-3">
    <Block className="h-[10px] opacity-50" text={label} style={{ width: 96 }} />
    <div className="flex flex-col gap-2">
      {configLines.map((width, i) => (
        <Block
          key={i}
          className={cn('h-[8px]', muted ? 'opacity-30' : 'opacity-60')}
          style={{ width }}
        />
      ))}
    </div>
  </div>
)

export const BranchCompare = () => {
  const { fromRunId = '', toRunId = '' } = useParams()
  const [selected, setSelected] = useState<IDiffEntry | undefined>()

  const sections = [...new Set(diffEntries.map((entry) => entry.section))]

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-3">
        <Block className="h-[32px]" title="Base" text={fromRunId} />
        <Block
          className="h-[12px] flex-none opacity-40"
          icon="ArrowRightIcon"
          iconSize={14}
        />
        <Block className="h-[32px]" title="Compare" text={toRunId} />
        <div className="ml-auto flex items-center gap-3">
          {(['added', 'removed', 'changed'] as const).map((kind) => (
            <Block
              key={kind}
              className="h-[14px] rounded-full opacity-70"
              text={`${diffEntries.filter((e) => e.kind === kind).length} ${kind}`}
              style={{ width: 90 }}
            />
          ))}
        </div>
      </div>

      {sections.map((section) => (
        <Panel key={section} title={section}>
          <div className="flex flex-col gap-1">
            {diffEntries
              .filter((entry) => entry.section === section)
              .map((entry) => (
                <div
                  key={entry.id}
                  title={`${kindLabel[entry.kind]} — ${entry.label}`}
                  onClick={() => setSelected(entry)}
                  className={`flex items-center gap-4 ${rowHoverClass}`}
                >
                  <Block
                    className="h-[14px] w-[72px] flex-none rounded-full opacity-70"
                    text={kindLabel[entry.kind]}
                  />
                  <div className="flex min-w-0 flex-1 flex-col gap-1.5">
                    <Block
                      className="h-[12px] cursor-pointer"
                      text={entry.label}
                      style={{ width: 150 }}
                    />
                    <Block
                      className="h-[8px] w-[96px] opacity-50"
                      text={entry.id}
                    />
                  </div>
                  <Block
                    className="h-[10px] flex-none opacity-50"
                    icon="CaretRightIcon"
                    iconSize={12}
                  />
                </div>
              ))}
          </div>
        </Panel>
      ))}

      {selected && (
        <Drawer title={selected.label} onClose={() => setSelected(undefined)}>
          <Block
            className="h-[14px] w-[80px] rounded-full opacity-70"
            text={kindLabel[selected.kind]}
          />

          <div className="flex gap-6">
            <DiffColumn label={fromRunId} muted={selected.kind === 'added'} />
            <DiffColumn label={toRunId} muted={selected.kind === 'removed'} />
          </div>
        </Drawer>
      )}
    </div>
  )
}
