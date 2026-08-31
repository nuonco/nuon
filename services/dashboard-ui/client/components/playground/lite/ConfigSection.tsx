import { useState } from 'react'
import { Block } from './Block'
import { Panel } from './Panel'
import { Toolbar } from './Toolbar'
import type { IBranchConfigItem } from './fixtures'
import { rowHoverClass } from './utils'

const PAGE_SIZE = 5

export interface IConfigSection {
  title: string
  items: IBranchConfigItem[]
  filters?: string[]
  onSelect: (item: IBranchConfigItem) => void
}

export const ConfigSection = ({
  title,
  items,
  filters,
  onSelect,
}: IConfigSection) => {
  const [visible, setVisible] = useState(PAGE_SIZE)
  const shown = items.slice(0, visible)
  const remaining = items.length - shown.length
  const isSearchable = items.length > PAGE_SIZE

  return (
    <Panel title={title} action="View source">
      {isSearchable && <Toolbar searchWidth={240} filters={filters ?? []} />}

      <div className="flex flex-col gap-1">
        {shown.map((item) => (
          <div
            key={item.id}
            className={`flex items-center gap-4 ${rowHoverClass}`}
            title={item.label}
            onClick={() => onSelect(item)}
          >
            <Block className="h-[10px] w-[10px] flex-none rounded-full" />
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              <Block
                className="h-[12px] cursor-pointer"
                text={item.label}
                style={{ width: 150 }}
              />
              <Block className="h-[8px] w-[96px] opacity-50" text={item.id} />
            </div>
            <Block className="h-[10px] w-[110px] flex-none opacity-50" />
            <Block
              className="h-[10px] flex-none opacity-50"
              icon="CaretRightIcon"
              iconSize={12}
            />
          </div>
        ))}
      </div>

      {remaining > 0 && (
        <Block
          className="h-[28px] w-full cursor-pointer opacity-60"
          title={`View ${Math.min(remaining, PAGE_SIZE)} more`}
          icon="CaretDownIcon"
          iconSize={12}
          text={`View more (${remaining})`}
          onClick={() => setVisible(visible + PAGE_SIZE)}
        />
      )}
    </Panel>
  )
}
