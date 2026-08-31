import { Block } from './Block'
import { Panel } from './Panel'
import { StatTile } from './StatTile'

const chartBars = [42, 68, 35, 80, 54, 72, 48, 92, 61, 38, 76, 55]

const activityRows = ['72%', '54%', '81%', '46%', '63%', '38%']

const installRows = ['58%', '44%', '66%', '38%', '52%']

const Chart = () => (
  <div className="flex items-end gap-2 h-[180px]">
    {chartBars.map((height, i) => (
      <Block key={i} className="flex-1" style={{ height: `${height}%` }} />
    ))}
  </div>
)

const ActivityFeed = () => (
  <div className="flex flex-col gap-4">
    {activityRows.map((width, i) => (
      <div key={i} className="flex items-center gap-3">
        <Block className="w-[24px] h-[24px] rounded-full flex-none" />
        <div className="flex flex-1 flex-col gap-1.5">
          <Block className="h-[10px]" style={{ width }} />
          <Block className="w-[40%] h-[8px] opacity-50" />
        </div>
      </div>
    ))}
  </div>
)

const InstallsTable = () => (
  <div className="flex flex-col gap-3">
    <div className="flex items-center gap-4 pb-1">
      <Block className="w-[180px] h-[8px] opacity-50" title="name" />
      <Block className="w-[120px] h-[8px] opacity-50" title="app" />
      <Block className="w-[64px] h-[8px] opacity-50 ml-auto" title="status" />
      <Block className="w-[72px] h-[8px] opacity-50" title="updated" />
    </div>
    {installRows.map((width, i) => (
      <div key={i} className="flex items-center gap-4 py-1">
        <Block className="h-[12px]" style={{ width }} />
        <Block className="w-[120px] h-[10px] opacity-60" />
        <Block className="w-[64px] h-[16px] rounded-full ml-auto" />
        <Block className="w-[72px] h-[10px] opacity-50" />
      </div>
    ))}
  </div>
)

export const HomePage = () => (
  <div className="flex flex-col gap-6">
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatTile label="Apps" valueWidth={40} />
      <StatTile label="Installs" valueWidth={56} />
      <StatTile label="Deploys today" valueWidth={72} />
      <StatTile label="Needs attention" valueWidth={32} />
    </div>

    <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <Panel
        title="Deploy activity"
        action="Last 30 days"
        className="lg:col-span-2"
      >
        <Chart />
      </Panel>

      <Panel title="Recent activity" action="View all">
        <ActivityFeed />
      </Panel>
    </div>

    <Panel title="Installs" action="View all">
      <InstallsTable />
    </Panel>
  </div>
)
