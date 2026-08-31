import { Block } from './Block'
import { Panel } from './Panel'
import { Toolbar } from './Toolbar'
import { inputGroups } from './fixtures'
import { rowHoverClass } from './utils'

export const InstallInputs = () => (
  <div className="flex flex-col gap-6">
    <Toolbar filters={['Group']} />

    {inputGroups.map((group) => (
      <Panel key={group.title} title={group.title} action="Edit">
        <div className="flex flex-col gap-1">
          {group.rows.map((row, i) => (
            <div
              key={i}
              className={`grid grid-cols-[minmax(0,1fr)_minmax(0,2fr)] gap-5 items-center ${rowHoverClass}`}
            >
              <Block
                className="h-[10px] max-w-full"
                style={{ width: row.key }}
              />
              <Block
                className="h-[10px] max-w-full opacity-60"
                style={{ width: row.value }}
              />
            </div>
          ))}
        </div>
      </Panel>
    ))}
  </div>
)
