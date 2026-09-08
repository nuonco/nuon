import { AppsTable } from '../components/organisms/AppsTable'
import { RouteScaffold } from '../components/organisms/RouteScaffold'
import { useBreadcrumbs } from '../hooks/use-breadcrumbs'
import { useOrg } from '../providers/org-provider'

export const Apps = () => {
  const { org, orgId } = useOrg()

  useBreadcrumbs([
    {
      label: org?.name,
      href: orgId ? `/${orgId}` : undefined,
      loadingWidth: 16,
    },
    { label: 'Apps' },
  ])

  return (
    <RouteScaffold
      title="Apps"
      description="Manage applications for this organization."
    >
      <AppsTable />
    </RouteScaffold>
  )
}
