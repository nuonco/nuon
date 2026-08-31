import { EntityPage, StatePanel } from './EntityPage'

export default {
  title: 'Playground/Lite/EntityPage',
}

export const Component = () => (
  <div className="p-4">
    <EntityPage
      stats={['Status', 'Version', 'Last deploy', 'Drift']}
      historyTitle="Deploys"
    />
  </div>
)

export const Stack = () => (
  <div className="p-4">
    <EntityPage
      stats={['Status', 'Roles', 'Last run', 'Drift']}
      historyTitle="Stack runs"
    >
      <StatePanel />
    </EntityPage>
  </div>
)
