import { AppsTable } from '../components/organisms/AppsTable'
import { RouteScaffold } from '../components/organisms/RouteScaffold'

export const Apps = () => (
  <RouteScaffold
    title="Apps"
    description="Manage applications for this organization."
  >
    <AppsTable />
  </RouteScaffold>
)
