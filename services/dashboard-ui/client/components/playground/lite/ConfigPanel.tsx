import { Block } from './Block'
import { Panel } from './Panel'
import { configLines } from './fixtures'
import { rowHoverClass } from './utils'

export interface IConfigPanel {
  title?: string
  action?: string
}

export const ConfigPanel = ({
  title = 'Configuration',
  action = 'Edit',
}: IConfigPanel) => (
  <Panel title={title} action={action}>
    <div className="flex flex-col gap-1">
      {configLines.map((line, i) => (
        <div
          key={i}
          className={`grid grid-cols-[minmax(0,1fr)_minmax(0,2fr)] gap-5 items-center ${rowHoverClass}`}
        >
          <Block className="h-[10px] max-w-full" style={{ width: line.key }} />
          <Block
            className="h-[10px] max-w-full opacity-60"
            style={{ width: line.value }}
          />
        </div>
      ))}
    </div>
  </Panel>
)
