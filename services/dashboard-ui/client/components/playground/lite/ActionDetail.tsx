import { useParams } from 'react-router'
import { ConfigPanel } from './ConfigPanel'
import { Panel } from './Panel'
import { RunHistory } from './RunHistory'
import { StatTile } from './StatTile'

export const ActionDetail = () => {
  const { installId = '', actionId = '' } = useParams()

  return (
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {['Status', 'Trigger', 'Last run', 'Runs'].map((stat) => (
          <StatTile key={stat} label={stat} />
        ))}
      </div>

      <ConfigPanel />

      <Panel title="Run history" action="View all">
        <RunHistory
          basePath={`/installs/${installId}/actions/${actionId}/runs`}
        />
      </Panel>
    </div>
  )
}
