import { InstallsTable } from '../components/organisms/InstallsTable'
import { RouteScaffold } from '../components/organisms/RouteScaffold'

export const Installs = () => (
  <RouteScaffold
    title="Installs"
    description="Manage installations for this organization."
  >
    <InstallsTable />
  </RouteScaffold>
)
