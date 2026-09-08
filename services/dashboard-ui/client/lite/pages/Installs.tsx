import { InstallsTable } from '../components/organisms/InstallsTable'
import { RouteScaffold } from '../components/organisms/RouteScaffold'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { useOrg } from '../providers/org-provider'

export const Installs = () => {
  const { org, orgId } = useOrg()

  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    { label: 'Installs' },
  ])

  return (
    <RouteScaffold
      title="Installs"
      description="Manage installations for this organization."
    >
      <InstallsTable />
    </RouteScaffold>
  )
}
