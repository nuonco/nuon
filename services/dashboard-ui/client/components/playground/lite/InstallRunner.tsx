import { Block } from './Block'
import { Panel } from './Panel'
import { StatTile } from './StatTile'

const logLines = [
  '84%',
  '62%',
  '91%',
  '48%',
  '76%',
  '58%',
  '88%',
  '44%',
  '70%',
  '95%',
  '52%',
  '80%',
]

export const InstallRunner = () => (
  <div className="flex flex-col gap-6">
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatTile label="Status" valueWidth={72} />
      <StatTile label="Version" valueWidth={64} />
      <StatTile label="Heartbeat" valueWidth={88} />
      <StatTile label="Jobs" valueWidth={40} />
    </div>

    <Panel title="Runner logs" action="Download">
      <div className="flex flex-col gap-2">
        {logLines.map((width, i) => (
          <Block key={i} className="h-[8px] opacity-60" style={{ width }} />
        ))}
      </div>
    </Panel>
  </div>
)
