import { useParams } from 'react-router'
import { ConfigPanel } from './ConfigPanel'
import { InstallDetail } from './InstallDetail'
import { InstallResources } from './InstallResources'
import { Panel } from './Panel'
import { RunHistory } from './RunHistory'
import { StatTile } from './StatTile'

export const ComponentDetailRoute = () => {
  const { installId = '', componentId = '' } = useParams()

  return (
    <InstallDetail
      crumbs={[{ label: componentId }]}
      actions={['Deploy']}
    >
      <div className="flex flex-col gap-6">
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
          {['Status', 'Version', 'Last deploy', 'Drift'].map((stat) => (
            <StatTile key={stat} label={stat} />
          ))}
        </div>

        <ConfigPanel />

        <Panel title="Resources" action="Refresh">
          <InstallResources />
        </Panel>

        <Panel title="Deploys" action="View all">
          <RunHistory
            basePath={`/installs/${installId}/components/${componentId}/deploys`}
          />
        </Panel>
      </div>
    </InstallDetail>
  )
}
