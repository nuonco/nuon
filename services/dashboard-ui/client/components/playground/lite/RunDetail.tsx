import { Block } from './Block'
import { ConfigPanel } from './ConfigPanel'
import { Panel } from './Panel'
import { PolicyChecks } from './PolicyChecks'
import { StatTile } from './StatTile'
import { outputRows, traceSpans } from './fixtures'
import { rowHoverClass } from './utils'

const logLines = [
  '84%', '62%', '91%', '48%', '76%', '58%', '88%', '44%',
  '70%', '95%', '52%', '80%', '66%', '38%', '72%',
]

const RunStats = () => (
  <div className="grid grid-cols-2 gap-4 lg:grid-cols-5">
    {['Status', 'Duration', 'Role', 'Triggered by', 'Started'].map((stat) => (
      <StatTile key={stat} label={stat} />
    ))}
  </div>
)

export const RunSummary = () => (
  <div className="flex flex-col gap-6">
    <RunStats />
    <PolicyChecks />
    <ConfigPanel title="Inputs" action="Copy" />
    <Panel title="Steps">
      <div className="flex flex-col gap-2">
        {['72%', '88%', '54%', '66%'].map((width, i) => (
          <div key={i} className={`flex items-center gap-4 ${rowHoverClass}`}>
            <Block className="h-[16px] w-[16px] flex-none rounded-full" />
            <Block className="h-[12px]" style={{ width }} />
            <Block className="h-[10px] w-[72px] flex-none opacity-50" />
          </div>
        ))}
      </div>
    </Panel>
  </div>
)

export const RunLogs = () => (
  <div className="flex flex-col gap-6">
    <RunStats />
    <Panel title="Logs" action="Download">
      <div className="flex flex-col gap-2">
        {logLines.map((width, i) => (
          <Block key={i} className="h-[8px] opacity-60" style={{ width }} />
        ))}
      </div>
    </Panel>
  </div>
)

export const RunTrace = () => (
  <div className="flex flex-col gap-6">
    <RunStats />
    <Panel title="Trace">
      <div className="flex flex-col gap-2">
        {traceSpans.map((span, i) => (
          <div key={i} className="flex items-center gap-3">
            <div style={{ width: span.indent * 24 }} />
            <Block className="h-[8px] w-[80px] flex-none opacity-50" />
            <Block
              className="h-[14px] rounded-full opacity-70"
              style={{ width: span.width }}
            />
          </div>
        ))}
      </div>
    </Panel>
  </div>
)

export const RunOutputs = () => (
  <div className="flex flex-col gap-6">
    <RunStats />
    <Panel title="Outputs" action="Copy">
      <div className="flex flex-col gap-1">
        {outputRows.map((row, i) => (
          <div
            key={i}
            className={`grid grid-cols-[minmax(0,1fr)_minmax(0,2fr)] gap-5 items-center ${rowHoverClass}`}
          >
            <Block className="h-[10px] max-w-full" style={{ width: row.key }} />
            <Block
              className="h-[10px] max-w-full opacity-60"
              style={{ width: row.value }}
            />
          </div>
        ))}
      </div>
    </Panel>
  </div>
)
