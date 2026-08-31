import { Block } from './Block'
import { Panel } from './Panel'
import { rowHoverClass } from './utils'

const labels = [
  { key: 116, value: 180 },
  { key: 92, value: 148 },
  { key: 140, value: 210 },
  { key: 104, value: 132 },
]

export const BranchLabels = () => (
  <Panel title="Labels" action="Edit">
    <div className="flex flex-col gap-1">
      {labels.map((label, i) => (
        <div
          key={i}
          className={`grid grid-cols-[minmax(0,1fr)_minmax(0,2fr)] gap-5 items-center ${rowHoverClass}`}
        >
          <Block className="h-[10px] max-w-full" style={{ width: label.key }} />
          <Block
            className="h-[14px] max-w-full rounded-full opacity-70"
            style={{ width: label.value }}
          />
        </div>
      ))}
    </div>
  </Panel>
)
